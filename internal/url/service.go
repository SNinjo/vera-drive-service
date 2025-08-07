package url

import (
	"strconv"

	"vera-identity-service/internal/apperror"
)

type Service interface {
	CreateURL(creates *RequestBody, userID int) error
	GetRootID(userID int) (string, error)
	GetURL(id string, userID int) (*URLResponse, error)
	ReplaceURL(id string, updates *RequestBody, userID int) error
	DeleteURL(id string, userID int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) validateOwnership(nodeID string, userID int) error {
	node, err := s.repo.GetOne(nodeID)
	if err != nil {
		return err
	}
	if node == nil {
		return apperror.New(apperror.CodeURLNotFound, "URL not found | id: "+nodeID)
	}
	if node.UserID != userID {
		return apperror.New(
			apperror.CodeURLAccessDenied, "Access denied | userID: "+strconv.Itoa(userID)+", nodeID: "+nodeID)
	}
	return nil
}

func (s *service) validateNameUniqueness(name string, parentID string, excludeID *string) error {
	siblings, err := s.repo.GetChildren(parentID)
	if err != nil {
		return err
	}
	for _, sibling := range siblings {
		if excludeID != nil && sibling.ID == *excludeID {
			continue
		}
		if sibling.Name == name {
			return apperror.New(apperror.CodeURLNameAlreadyExists, "Name already exists in this folder | name: "+name+", parentID: "+parentID)
		}
	}
	return nil
}

func (s *service) GetRootID(userID int) (string, error) {
	root, err := s.repo.GetRoot(userID)
	if err != nil {
		return "", err
	}
	if root == nil {
		root = &URLNode{
			UserID: userID,
			Name:   "",
			Type:   "folder",
		}
		if err := s.repo.Create(root); err != nil {
			return "", err
		}
	}
	return root.ID, nil
}

func (s *service) GetURL(id string, userID int) (*URLResponse, error) {
	if err := s.validateOwnership(id, userID); err != nil {
		return nil, err
	}

	node, err := s.repo.GetOne(id)
	if err != nil {
		return nil, err
	}

	parents, err := s.repo.GetParentUpToRoot(id)
	if err != nil {
		return nil, err
	}

	children, err := s.repo.GetChildren(id)
	if err != nil {
		return nil, err
	}

	return newURLResponse(node, parents, children), nil
}

func (s *service) CreateURL(creates *RequestBody, userID int) error {
	if err := s.validateOwnership(creates.ParentID, userID); err != nil {
		return err
	}
	if err := s.validateNameUniqueness(creates.Name, creates.ParentID, nil); err != nil {
		return err
	}

	node := &URLNode{
		UserID:   userID,
		ParentID: &creates.ParentID,
		Name:     creates.Name,
		Type:     creates.Type,
		URL:      creates.URL,
	}
	if err := s.repo.Create(node); err != nil {
		return err
	}

	return nil
}

func (s *service) ReplaceURL(id string, updates *RequestBody, userID int) error {
	if err := s.validateOwnership(id, userID); err != nil {
		return err
	}
	if err := s.validateOwnership(updates.ParentID, userID); err != nil {
		return err
	}
	if err := s.validateNameUniqueness(updates.Name, updates.ParentID, &id); err != nil {
		return err
	}

	node, err := s.repo.GetOne(id)
	if err != nil {
		return err
	}
	if node == nil {
		return apperror.New(apperror.CodeURLNotFound, "URL not found | id: "+id)
	}

	node.ParentID = &updates.ParentID
	node.Name = updates.Name
	node.Type = updates.Type
	node.URL = updates.URL

	err = s.repo.Update(node)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteURL(id string, userID int) error {
	if err := s.validateOwnership(id, userID); err != nil {
		return err
	}

	if err := s.repo.SoftDelete(id); err != nil {
		return err
	}

	return nil
}
