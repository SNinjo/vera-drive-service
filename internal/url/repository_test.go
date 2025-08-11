package url

import (
	"log"
	"os"
	"testing"
	"time"

	"vera-identity-service/internal/config"
	"vera-identity-service/internal/db"
	"vera-identity-service/test"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var d *gorm.DB

func TestMain(m *testing.M) {
	// Setup
	dbURL, closeDB, err := test.SetupPostgresql()
	if err != nil {
		log.Fatal(err)
	}

	d, err = db.NewDatabase(&config.Config{DatabaseURL: dbURL})
	if err != nil {
		log.Fatal(err)
	}

	err = d.AutoMigrate(&URLNode{})
	if err != nil {
		log.Fatal(err)
	}

	// Run
	code := m.Run()

	// Teardown
	closeDB()

	os.Exit(code)
}

func TestRepository_NewRepository_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)

	// Act
	repo := NewRepository(d)

	// Assert
	assert.NotNil(t, repo)
	assert.IsType(t, &repository{}, repo)
}

func TestRepository_Create_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)
	node := &URLNode{
		UserID:   1,
		ParentID: nil,
		Name:     "name",
		Type:     "folder",
		URL:      nil,
	}

	// Act
	err = repo.Create(node)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, node.ID, node.ID)
	assert.Equal(t, 1, node.UserID)
	assert.Nil(t, node.ParentID)
	assert.Equal(t, "name", node.Name)
	assert.Equal(t, "folder", node.Type)
	assert.Nil(t, node.URL)
	assert.WithinDuration(t, time.Now().UTC(), node.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), node.UpdatedAt, time.Second)
	assert.Nil(t, node.DeletedAt)

	savedNode, err := repo.GetOne(node.ID)
	require.NoError(t, err)
	assert.Equal(t, node.ID, savedNode.ID)
	assert.Equal(t, 1, savedNode.UserID)
	assert.Nil(t, savedNode.ParentID)
	assert.Equal(t, "name", savedNode.Name)
	assert.Equal(t, "folder", savedNode.Type)
	assert.Nil(t, savedNode.URL)
	assert.WithinDuration(t, time.Now().UTC(), savedNode.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), savedNode.UpdatedAt, time.Second)
	assert.Nil(t, savedNode.DeletedAt)
}
func TestRepository_Create_SpecificID(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)
	specificID := uuid.New().String()
	node := &URLNode{
		ID:       specificID,
		UserID:   1,
		ParentID: nil,
		Name:     "name",
		Type:     "url",
		URL:      test.StringPtr("https://example.com"),
	}

	// Act
	err = repo.Create(node)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, specificID, node.ID)

	savedNode, err := repo.GetOne(node.ID)
	require.NoError(t, err)
	assert.NotNil(t, savedNode)
}
func TestRepository_Create_DuplicateID(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	duplicateID := uuid.New().String()
	node := &URLNode{
		ID:       duplicateID,
		UserID:   1,
		ParentID: nil,
		Name:     "name",
		Type:     "url",
		URL:      test.StringPtr("https://example.com"),
	}
	err = d.Create(node).Error
	require.NoError(t, err)

	// Act
	err = repo.Create(node)

	// Assert
	assert.Error(t, err)
}

func TestRepository_GetRoot_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	root := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: nil,
		Name:     "root",
		Type:     "folder",
	}
	err = d.Create(root).Error
	require.NoError(t, err)

	// Act
	result, err := repo.GetRoot(1)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, root.ID, result.ID)
	assert.Equal(t, root.UserID, result.UserID)
	assert.Nil(t, result.ParentID)
	assert.Equal(t, "root", result.Name)
	assert.Equal(t, "folder", result.Type)
	assert.Nil(t, result.URL)
	assert.WithinDuration(t, time.Now().UTC(), result.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), result.UpdatedAt, time.Second)
	assert.Nil(t, result.DeletedAt)
}
func TestRepository_GetRoot_NonExistentUser(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	// Act
	result, err := repo.GetRoot(-1)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRepository_GetOne_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	node := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: nil,
		Name:     "name",
		Type:     "url",
		URL:      test.StringPtr("https://example.com"),
	}
	err = d.Create(node).Error
	require.NoError(t, err)

	// Act
	result, err := repo.GetOne(node.ID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, node.ID, result.ID)
	assert.Equal(t, node.UserID, result.UserID)
	assert.Equal(t, node.ParentID, result.ParentID)
	assert.Equal(t, node.Name, result.Name)
	assert.Equal(t, node.Type, result.Type)
	assert.Equal(t, node.URL, result.URL)
	assert.WithinDuration(t, time.Now().UTC(), result.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), result.UpdatedAt, time.Second)
	assert.Nil(t, result.DeletedAt)
}
func TestRepository_GetOne_NonExistentNode(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	nonExistentID := uuid.New().String()

	// Act
	result, err := repo.GetOne(nonExistentID)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result)
}
func TestRepository_GetOne_FilterSoftDeleted(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	node := &URLNode{
		ID:        uuid.New().String(),
		UserID:    1,
		ParentID:  nil,
		Name:      "name",
		Type:      "url",
		URL:       test.StringPtr("https://example.com"),
		DeletedAt: &time.Time{},
	}
	err = d.Create(node).Error
	require.NoError(t, err)

	// Act
	result, err := repo.GetOne(node.ID)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestRepository_GetParentUpToRoot_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	root := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: nil,
		Name:     "Root",
		Type:     "folder",
	}
	err = d.Create(root).Error
	require.NoError(t, err)

	parent := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: &root.ID,
		Name:     "Parent",
		Type:     "folder",
	}
	err = d.Create(parent).Error
	require.NoError(t, err)

	child := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: &parent.ID,
		Name:     "Child",
		Type:     "folder",
	}
	err = d.Create(child).Error
	require.NoError(t, err)

	// Act
	parents, err := repo.GetParentUpToRoot(child.ID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, parents, 2)

	assert.Equal(t, root.ID, parents[0].ID)
	assert.Equal(t, root.UserID, parents[0].UserID)
	assert.Nil(t, parents[0].ParentID)
	assert.Equal(t, "Root", parents[0].Name)
	assert.Equal(t, "folder", parents[0].Type)
	assert.Nil(t, parents[0].URL)
	assert.WithinDuration(t, time.Now().UTC(), parents[0].CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), parents[0].UpdatedAt, time.Second)
	assert.Nil(t, parents[0].DeletedAt)

	assert.Equal(t, parent.ID, parents[1].ID)
	assert.Equal(t, parent.UserID, parents[1].UserID)
	assert.Equal(t, root.ID, *parents[1].ParentID)
	assert.Equal(t, "Parent", parents[1].Name)
	assert.Equal(t, "folder", parents[1].Type)
	assert.Nil(t, parents[1].URL)
	assert.WithinDuration(t, time.Now().UTC(), parents[1].CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), parents[1].UpdatedAt, time.Second)
	assert.Nil(t, parents[1].DeletedAt)
}
func TestRepository_GetParentUpToRoot_NoParent(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	root := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: nil,
		Name:     "Root",
		Type:     "folder",
	}
	err = d.Create(root).Error
	require.NoError(t, err)

	// Act
	parents, err := repo.GetParentUpToRoot(root.ID)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, parents)
}
func TestRepository_GetParentUpToRoot_NonExistentNode(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	nonExistentID := uuid.New().String()

	// Act
	parents, err := repo.GetParentUpToRoot(nonExistentID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, parents)
}
func TestRepository_GetParentUpToRoot_FilterSoftDeleted(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	parent := &URLNode{
		ID:     uuid.New().String(),
		UserID: 1,
		Name:   "Parent",
		Type:   "folder",
	}
	err = d.Create(parent).Error
	require.NoError(t, err)

	child := &URLNode{
		ID:        uuid.New().String(),
		UserID:    1,
		ParentID:  &parent.ID,
		Name:      "Child",
		Type:      "folder",
		DeletedAt: &time.Time{},
	}
	err = d.Create(child).Error
	require.NoError(t, err)

	// Act
	parents, err := repo.GetParentUpToRoot(child.ID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, parents)
}

func TestRepository_GetChildren_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	parent := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: nil,
		Name:     "parent",
		Type:     "folder",
	}
	err = d.Create(parent).Error
	require.NoError(t, err)

	child1 := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: &parent.ID,
		Name:     "child1",
		Type:     "folder",
	}
	child2 := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: &parent.ID,
		Name:     "child2",
		Type:     "url",
		URL:      test.StringPtr("https://example.com"),
	}
	err = d.Create(child1).Error
	require.NoError(t, err)
	err = d.Create(child2).Error
	require.NoError(t, err)

	// Act
	children, err := repo.GetChildren(parent.ID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, children, 2)

	assert.Equal(t, child1.ID, children[0].ID)
	assert.Equal(t, child1.UserID, children[0].UserID)
	assert.Equal(t, parent.ID, *children[0].ParentID)
	assert.Equal(t, "child1", children[0].Name)
	assert.Equal(t, "folder", children[0].Type)
	assert.Nil(t, children[0].URL)
	assert.WithinDuration(t, time.Now().UTC(), children[0].CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), children[0].UpdatedAt, time.Second)
	assert.Nil(t, children[0].DeletedAt)

	assert.Equal(t, child2.ID, children[1].ID)
	assert.Equal(t, child2.UserID, children[1].UserID)
	assert.Equal(t, parent.ID, *children[1].ParentID)
	assert.Equal(t, "child2", children[1].Name)
	assert.Equal(t, "url", children[1].Type)
	assert.Equal(t, "https://example.com", *children[1].URL)
	assert.WithinDuration(t, time.Now().UTC(), children[1].CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), children[1].UpdatedAt, time.Second)
	assert.Nil(t, children[1].DeletedAt)
}
func TestRepository_GetChildren_NoChildren(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	node := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: nil,
		Name:     "node",
		Type:     "folder",
	}
	err = d.Create(node).Error
	require.NoError(t, err)

	// Act
	children, err := repo.GetChildren(node.ID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, children)
}
func TestRepository_GetChildren_NonExistentNode(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	nonExistentID := uuid.New().String()

	// Act
	children, err := repo.GetChildren(nonExistentID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, children)
}
func TestRepository_GetChildren_FilterSoftDeleted(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	parent := &URLNode{
		ID:       uuid.New().String(),
		UserID:   1,
		ParentID: nil,
		Name:     "parent",
		Type:     "folder",
	}
	err = d.Create(parent).Error
	require.NoError(t, err)

	child := &URLNode{
		ID:        uuid.New().String(),
		UserID:    1,
		ParentID:  &parent.ID,
		Name:      "child",
		Type:      "folder",
		DeletedAt: &time.Time{},
	}
	err = d.Create(child).Error
	require.NoError(t, err)

	// Act
	children, err := repo.GetChildren(parent.ID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, children)
}

func TestRepository_Update_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	node := &URLNode{
		UserID: 1,
		Name:   "name",
		Type:   "folder",
	}
	err = d.Create(node).Error
	require.NoError(t, err)

	originalUpdatedAt := node.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	node.Name = "new name"
	node.Type = "url"
	url := "https://example.com"
	node.URL = &url

	// Act
	err = repo.Update(node)

	// Assert
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().UTC(), node.UpdatedAt, time.Second)

	updated, err := repo.GetOne(node.ID)
	require.NoError(t, err)
	assert.Equal(t, "new name", updated.Name)
	assert.Equal(t, "url", updated.Type)
	assert.Equal(t, "https://example.com", *updated.URL)
	assert.True(t, updated.UpdatedAt.After(originalUpdatedAt))
}
func TestRepository_Update_NonExistentNode(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	node := &URLNode{
		ID:     uuid.New().String(),
		UserID: 1,
		Name:   "non-existent",
		Type:   "folder",
	}

	// Act
	err = repo.Update(node)

	// Assert
	require.NoError(t, err)

	updated, err := repo.GetOne(node.ID)
	require.NoError(t, err)
	assert.Equal(t, "non-existent", updated.Name)
	assert.Equal(t, "folder", updated.Type)
	assert.Nil(t, updated.URL)
	assert.WithinDuration(t, time.Now().UTC(), updated.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), updated.UpdatedAt, time.Second)
	assert.Nil(t, updated.DeletedAt)
}

func TestRepository_SoftDelete_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	node := &URLNode{
		UserID: 1,
		Name:   "name",
		Type:   "folder",
	}
	err = d.Create(node).Error
	require.NoError(t, err)

	// Act
	err = repo.SoftDelete(node.ID)

	// Assert
	require.NoError(t, err)

	deleted := &URLNode{}
	err = d.Unscoped().Where("id = ?", node.ID).First(deleted).Error
	require.NoError(t, err)
	assert.Equal(t, "name", deleted.Name)
	assert.Equal(t, "folder", deleted.Type)
	assert.Nil(t, deleted.URL)
	assert.WithinDuration(t, time.Now().UTC(), deleted.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), deleted.UpdatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC(), *deleted.DeletedAt, time.Second)
}
func TestRepository_SoftDelete_NonExistentNode(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	nonExistentID := uuid.New().String()

	// Act
	err = repo.SoftDelete(nonExistentID)

	// Assert
	require.NoError(t, err)
}
