package url

import (
	"testing"

	"github.com/vera/vera-drive-service/internal/apperror"
	"github.com/vera/vera-drive-service/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(node *URLNode) error {
	args := m.Called(node)
	return args.Error(0)
}
func (m *MockRepository) GetRoot(userID int) (*URLNode, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*URLNode), args.Error(1)
}
func (m *MockRepository) GetOne(id string) (*URLNode, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*URLNode), args.Error(1)
}
func (m *MockRepository) GetParentUpToRoot(id string) ([]URLNode, error) {
	args := m.Called(id)
	return args.Get(0).([]URLNode), args.Error(1)
}
func (m *MockRepository) GetChildren(id string) ([]URLNode, error) {
	args := m.Called(id)
	return args.Get(0).([]URLNode), args.Error(1)
}
func (m *MockRepository) Update(node *URLNode) error {
	args := m.Called(node)
	return args.Error(0)
}
func (m *MockRepository) SoftDelete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestService_NewService_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}

	// Act
	s := NewService(mockRepo)

	// Assert
	assert.IsType(t, &service{}, s)
	assert.Equal(t, mockRepo, s.(*service).repo)
}

func TestService_validateOwnership_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1

	expectedNode := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "test-node",
		Type:   "folder",
	}

	mockRepo.On("GetOne", nodeID).Return(expectedNode, nil)

	// Act
	err := service.validateOwnership(nodeID, userID)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_validateOwnership_NodeNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1

	mockRepo.On("GetOne", nodeID).Return(nil, nil)

	// Act
	err := service.validateOwnership(nodeID, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeURLNotFound, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_validateOwnership_AccessDenied(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1
	ownerID := 2

	expectedNode := &URLNode{
		ID:     nodeID,
		UserID: ownerID,
		Name:   "test-node",
		Type:   "folder",
	}

	mockRepo.On("GetOne", nodeID).Return(expectedNode, nil)

	// Act
	err := service.validateOwnership(nodeID, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeURLAccessDenied, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_validateOwnership_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1

	mockRepo.On("GetOne", nodeID).Return(nil, assert.AnError)

	// Act
	err := service.validateOwnership(nodeID, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_validateNameUniqueness_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	name := "unique-name"
	parentID := "parent-id"
	siblings := []URLNode{
		{ID: "sibling1", Name: "existing-name", Type: "folder"},
	}

	mockRepo.On("GetChildren", parentID).Return(siblings, nil)

	// Act
	err := service.validateNameUniqueness(name, parentID, nil)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_validateNameUniqueness_ExcludeID(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	name := "existing-name"
	parentID := "parent-id"
	siblings := []URLNode{
		{ID: "sibling1", Name: name, Type: "folder"},
	}

	mockRepo.On("GetChildren", parentID).Return(siblings, nil)

	// Act
	err := service.validateNameUniqueness(name, parentID, test.StringPtr("sibling1"))

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_validateNameUniqueness_NameAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	name := "existing-name"
	parentID := "parent-id"
	siblings := []URLNode{
		{ID: "sibling1", Name: "existing-name", Type: "folder"},
	}

	mockRepo.On("GetChildren", parentID).Return(siblings, nil)

	// Act
	err := service.validateNameUniqueness(name, parentID, nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeURLNameAlreadyExists, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_validateNameUniqueness_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	name := "test-name"
	parentID := "parent-id"

	mockRepo.On("GetChildren", parentID).Return([]URLNode{}, assert.AnError)

	// Act
	err := service.validateNameUniqueness(name, parentID, nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetRootID_ExistingRoot(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	userID := 1
	expectedRoot := &URLNode{
		ID:     "root-id",
		UserID: userID,
		Name:   "root",
		Type:   "folder",
	}

	mockRepo.On("GetRoot", userID).Return(expectedRoot, nil)

	// Act
	rootID, err := service.GetRootID(userID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "root-id", rootID)
	mockRepo.AssertExpectations(t)
}
func TestService_GetRootID_NonExistentRoot(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	userID := 1
	creates := &URLNode{
		UserID: userID,
		Name:   "",
		Type:   "folder",
	}

	mockRepo.On("GetRoot", userID).Return(nil, nil)
	mockRepo.On("Create", creates).
		Run(func(args mock.Arguments) {
			node := args.Get(0).(*URLNode)
			node.ID = "mock-new-node-id"
		}).
		Return(nil)

	// Act
	rootID, err := service.GetRootID(userID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "mock-new-node-id", rootID)
	mockRepo.AssertExpectations(t)
}
func TestService_GetRootID_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	userID := 1

	mockRepo.On("GetRoot", userID).Return(nil, assert.AnError)

	// Act
	rootID, err := service.GetRootID(userID)

	// Assert
	assert.Empty(t, rootID)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}
func TestService_GetRootID_CreateRootError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	userID := 1

	mockRepo.On("GetRoot", userID).Return(nil, nil)
	mockRepo.On("Create", mock.AnythingOfType("*url.URLNode")).Return(assert.AnError)

	// Act
	rootID, err := service.GetRootID(userID)

	// Assert
	assert.Empty(t, rootID)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetURL_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1
	node := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "mock-node",
		Type:   "folder",
	}
	parents := []URLNode{
		{ID: "parent1", Name: "parent1", Type: "folder"},
		{ID: "parent2", Name: "parent2", Type: "folder"},
	}
	children := []URLNode{
		{ID: "child1", Name: "child1", Type: "url"},
		{ID: "child2", Name: "child2", Type: "folder"},
	}

	mockRepo.On("GetOne", nodeID).Return(node, nil)
	mockRepo.On("GetParentUpToRoot", nodeID).Return(parents, nil)
	mockRepo.On("GetChildren", nodeID).Return(children, nil)

	// Act
	response, err := service.GetURL(nodeID, userID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, nodeID, response.ID)
	assert.Equal(t, "mock-node", response.Name)
	assert.Equal(t, "folder", response.Type)

	assert.Len(t, response.Parent, 2)
	assert.Equal(t, "parent1", response.Parent[0].ID)
	assert.Equal(t, "parent2", response.Parent[1].ID)

	assert.Len(t, response.Children, 2)
	assert.Equal(t, "child1", response.Children[0].ID)
	assert.Equal(t, "child2", response.Children[1].ID)

	mockRepo.AssertExpectations(t)
}
func TestService_GetURL_OwnershipError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1

	mockRepo.On("GetOne", nodeID).Return(nil, nil)

	// Act
	response, err := service.GetURL(nodeID, userID)

	// Assert
	assert.Nil(t, response)
	assert.Equal(t, apperror.CodeURLNotFound, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_GetURL_GetOneError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1
	node := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "test-node",
		Type:   "folder",
	}

	mockRepo.On("GetOne", nodeID).Return(node, nil).Once()
	mockRepo.On("GetOne", nodeID).Return(nil, assert.AnError).Once()

	// Act
	response, err := service.GetURL(nodeID, userID)

	// Assert
	assert.Nil(t, response)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}
func TestService_GetURL_GetParentUpToRootError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1
	node := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "test-node",
		Type:   "folder",
	}

	mockRepo.On("GetOne", nodeID).Return(node, nil)
	mockRepo.On("GetParentUpToRoot", nodeID).Return([]URLNode{}, assert.AnError)

	// Act
	response, err := service.GetURL(nodeID, userID)

	// Assert
	assert.Nil(t, response)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}
func TestService_GetURL_GetChildrenError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1
	node := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "test-node",
		Type:   "folder",
	}

	mockRepo.On("GetOne", nodeID).Return(node, nil)
	mockRepo.On("GetParentUpToRoot", nodeID).Return([]URLNode{}, nil)
	mockRepo.On("GetChildren", nodeID).Return([]URLNode{}, assert.AnError)

	// Act
	response, err := service.GetURL(nodeID, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_CreateURL_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	userID := 1
	creates := &RequestBody{
		ParentID: "parent-id",
		Name:     "new-url",
		Type:     "url",
		URL:      test.StringPtr("https://example.com"),
	}
	createdNode := &URLNode{
		UserID:   userID,
		ParentID: test.StringPtr(creates.ParentID),
		Name:     creates.Name,
		Type:     creates.Type,
		URL:      creates.URL,
	}
	parentNode := &URLNode{
		ID:     "parent-id",
		UserID: userID,
		Name:   "parent",
		Type:   "folder",
	}

	mockRepo.On("GetOne", "parent-id").Return(parentNode, nil)
	mockRepo.On("GetChildren", "parent-id").Return([]URLNode{}, nil)
	mockRepo.On("Create", createdNode).Return(nil)

	// Act
	err := service.CreateURL(creates, userID)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_CreateURL_ParentOwnershipError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	userID := 1

	createReq := &RequestBody{
		ParentID: "parent-id",
		Name:     "new-url",
		Type:     "url",
		URL:      test.StringPtr("https://example.com"),
	}

	mockRepo.On("GetOne", "parent-id").Return(nil, nil)

	// Act
	err := service.CreateURL(createReq, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeURLNotFound, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_CreateURL_NameUniquenessError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	userID := 1
	creates := &RequestBody{
		ParentID: "parent-id",
		Name:     "existing-name",
		Type:     "url",
		URL:      test.StringPtr("https://example.com"),
	}
	parentNode := &URLNode{
		ID:     "parent-id",
		UserID: userID,
		Name:   "parent",
		Type:   "folder",
	}
	siblings := []URLNode{
		{ID: "sibling1", Name: "existing-name", Type: "url"},
	}

	mockRepo.On("GetOne", "parent-id").Return(parentNode, nil)
	mockRepo.On("GetChildren", "parent-id").Return(siblings, nil)

	// Act
	err := service.CreateURL(creates, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeURLNameAlreadyExists, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_CreateURL_CreateError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	userID := 1
	creates := &RequestBody{
		ParentID: "parent-id",
		Name:     "new-url",
		Type:     "url",
		URL:      test.StringPtr("https://example.com"),
	}
	parentNode := &URLNode{
		ID:     "parent-id",
		UserID: userID,
		Name:   "parent",
		Type:   "folder",
	}

	mockRepo.On("GetOne", "parent-id").Return(parentNode, nil)
	mockRepo.On("GetChildren", "parent-id").Return([]URLNode{}, nil)
	mockRepo.On("Create", mock.AnythingOfType("*url.URLNode")).Return(assert.AnError)

	// Act
	err := service.CreateURL(creates, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_ReplaceURL_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	newParentID := "new-parent-id"
	userID := 1
	updates := &RequestBody{
		ParentID: newParentID,
		Name:     "updated-name",
		Type:     "url",
		URL:      test.StringPtr("https://updated.com"),
	}
	node := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "old-name",
		Type:   "folder",
		URL:    nil,
	}
	updatedNode := &URLNode{
		ID:       nodeID,
		UserID:   userID,
		ParentID: test.StringPtr(newParentID),
		Name:     updates.Name,
		Type:     updates.Type,
		URL:      updates.URL,
	}
	newParentNode := &URLNode{
		ID:     newParentID,
		UserID: userID,
		Name:   "new-parent",
		Type:   "folder",
	}

	mockRepo.On("GetOne", nodeID).Return(node, nil).Once()
	mockRepo.On("GetOne", newParentID).Return(newParentNode, nil).Once()
	mockRepo.On("GetChildren", newParentID).Return([]URLNode{}, nil)
	mockRepo.On("GetOne", nodeID).Return(node, nil).Once()
	mockRepo.On("Update", updatedNode).Return(nil)

	// Act
	err := service.ReplaceURL(nodeID, updates, userID)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_ReplaceURL_UpdatedNodeOwnershipError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1
	updates := &RequestBody{
		ParentID: "new-parent-id",
		Name:     "updated-name",
		Type:     "url",
		URL:      test.StringPtr("https://updated.com"),
	}

	mockRepo.On("GetOne", nodeID).Return(nil, nil)

	// Act
	err := service.ReplaceURL(nodeID, updates, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeURLNotFound, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_ReplaceURL_NewParentOwnershipError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	newParentID := "new-parent-id"
	userID := 1
	node := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "old-name",
		Type:   "url",
		URL:    test.StringPtr("https://old.com"),
	}
	updates := &RequestBody{
		ParentID: newParentID,
		Name:     "updated-name",
		Type:     "url",
		URL:      test.StringPtr("https://updated.com"),
	}

	mockRepo.On("GetOne", nodeID).Return(node, nil)
	mockRepo.On("GetOne", newParentID).Return(nil, nil)

	// Act
	err := service.ReplaceURL(nodeID, updates, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeURLNotFound, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_ReplaceURL_NameUniquenessError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	newParentID := "new-parent-id"
	userID := 1
	updates := &RequestBody{
		ParentID: newParentID,
		Name:     "existing-name",
		Type:     "url",
		URL:      test.StringPtr("https://updated.com"),
	}
	node := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "old-name",
		Type:   "url",
		URL:    test.StringPtr("https://old.com"),
	}
	newParentNode := &URLNode{
		ID:     newParentID,
		UserID: userID,
		Name:   "new-parent",
		Type:   "folder",
	}
	siblings := []URLNode{
		{ID: "sibling1", Name: "existing-name", Type: "url"},
	}

	mockRepo.On("GetOne", nodeID).Return(node, nil)
	mockRepo.On("GetOne", nodeID).Return(node, nil)
	mockRepo.On("GetOne", newParentID).Return(newParentNode, nil)
	mockRepo.On("GetChildren", newParentID).Return(siblings, nil)

	// Act
	err := service.ReplaceURL(nodeID, updates, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeURLNameAlreadyExists, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_ReplaceURL_UpdateError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	newParentID := "new-parent-id"
	userID := 1
	updates := &RequestBody{
		ParentID: newParentID,
		Name:     "updated-name",
		Type:     "url",
		URL:      test.StringPtr("https://updated.com"),
	}
	existingNode := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "old-name",
		Type:   "url",
		URL:    test.StringPtr("https://old.com"),
	}
	newParentNode := &URLNode{
		ID:     newParentID,
		UserID: userID,
		Name:   "new-parent",
		Type:   "folder",
	}

	mockRepo.On("GetOne", nodeID).Return(existingNode, nil)
	mockRepo.On("GetOne", nodeID).Return(existingNode, nil)
	mockRepo.On("GetOne", newParentID).Return(newParentNode, nil)
	mockRepo.On("GetChildren", newParentID).Return([]URLNode{}, nil)
	mockRepo.On("Update", mock.AnythingOfType("*url.URLNode")).Return(assert.AnError)

	// Act
	err := service.ReplaceURL(nodeID, updates, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteURL_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1
	node := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "test-node",
		Type:   "url",
	}

	mockRepo.On("GetOne", nodeID).Return(node, nil)
	mockRepo.On("SoftDelete", nodeID).Return(nil)

	// Act
	err := service.DeleteURL(nodeID, userID)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_DeleteURL_OwnershipError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1

	mockRepo.On("GetOne", nodeID).Return(nil, nil)

	// Act
	err := service.DeleteURL(nodeID, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeURLNotFound, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_DeleteURL_SoftDeleteError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}
	nodeID := "mock-node-id"
	userID := 1
	node := &URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "test-node",
		Type:   "url",
	}

	mockRepo.On("GetOne", nodeID).Return(node, nil)
	mockRepo.On("SoftDelete", nodeID).Return(assert.AnError)

	// Act
	err := service.DeleteURL(nodeID, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}
