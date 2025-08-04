package url

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type URLNode struct {
	ID        string    `gorm:"type:uuid;primary_key"`
	UserID    int       `gorm:"type:int;index"`
	ParentID  *string   `gorm:"type:uuid;index"`
	Parent    *URLNode  `gorm:"foreignKey:ParentID"`
	Children  []URLNode `gorm:"foreignKey:ParentID"`
	Name      string    `gorm:"size:255;not null"`
	Type      string    `gorm:"size:10;not null;check:type IN ('folder','url')"`
	URL       *string   `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

func (URLNode) TableName() string {
	return "url_nodes"
}

func (u *URLNode) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *URLNode) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now().UTC()
	return nil
}

type Repository interface {
	Create(node *URLNode) error
	GetRoot(userID int) (*URLNode, error)
	GetOne(id string) (*URLNode, error)
	GetParentUpToRoot(id string) ([]URLNode, error)
	GetChildren(id string) ([]URLNode, error)
	Update(node *URLNode) error
	SoftDelete(id string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(node *URLNode) error {
	return r.db.Create(node).Error
}

func (r *repository) GetRoot(userID int) (*URLNode, error) {
	var root URLNode
	err := r.db.Where("parent_id IS NULL AND deleted_at IS NULL AND user_id = ?", userID).First(&root).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &root, nil
}

func (r *repository) GetOne(id string) (*URLNode, error) {
	var node URLNode
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&node).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *repository) GetParentUpToRoot(id string) ([]URLNode, error) {
	var parents []URLNode
	current, err := r.GetOne(id)
	if err != nil {
		return nil, err
	}

	for current != nil && current.ParentID != nil {
		parent, err := r.GetOne(*current.ParentID)
		if err != nil {
			return nil, err
		}

		parents = append([]URLNode{*parent}, parents...)
		current = parent
	}

	return parents, nil
}

func (r *repository) GetChildren(id string) ([]URLNode, error) {
	var children []URLNode
	err := r.db.Where("parent_id = ? AND deleted_at IS NULL", id).Find(&children).Error
	return children, err
}

func (r *repository) Update(node *URLNode) error {
	return r.db.Save(node).Error
}

func (r *repository) SoftDelete(id string) error {
	now := time.Now().UTC()
	return r.db.Model(&URLNode{}).Where("id = ?", id).Update("deleted_at", now).Error
}
