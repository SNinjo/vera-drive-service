package url

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
	"vera-identity-service/test"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateURL(creates *RequestBody, userID int) error {
	args := m.Called(creates, userID)
	return args.Error(0)
}
func (m *MockService) GetRootID(userID int) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}
func (m *MockService) GetURL(id string, userID int) (*URLResponse, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*URLResponse), args.Error(1)
}
func (m *MockService) ReplaceURL(id string, updates *RequestBody, userID int) error {
	args := m.Called(id, updates, userID)
	return args.Error(0)
}
func (m *MockService) DeleteURL(id string, userID int) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func TestHandler_NewHandler_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}

	// Act
	h := NewHandler(mockService)

	// Assert
	assert.IsType(t, &Handler{}, h)
	assert.Equal(t, mockService, h.service)
}

func TestHandler_CreateURL_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	userID := 1
	requestBody := RequestBody{
		ParentID: "123e4567-e89b-12d3-a456-426614174001",
		Name:     "folder",
		Type:     "folder",
	}
	requestJSON, _ := json.Marshal(requestBody)

	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestJSON))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", userID)

	mockService.On("CreateURL", &requestBody, userID).Return(nil)

	// Act
	handler.CreateURL(c)
	c.Writer.WriteHeaderNow()

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
func TestHandler_CreateURL_InvalidRequestBody(t *testing.T) {
	tests := []struct {
		name          string
		payload       string
		errorContains string
	}{
		{
			name:          "missing parent_id",
			payload:       `{"name": "folder", "type": "folder"}`,
			errorContains: "ParentID",
		},
		{
			name:          "invalid parent_id format",
			payload:       `{"parent_id": "not-a-uuid", "name": "folder", "type": "folder"}`,
			errorContains: "uuid",
		},
		{
			name:          "missing name",
			payload:       `{"parent_id": "123e4567-e89b-12d3-a456-426614174000", "type": "folder"}`,
			errorContains: "Name",
		},
		{
			name:          "name too long",
			payload:       `{"parent_id": "123e4567-e89b-12d3-a456-426614174000", "name": "this name is definitely too long", "type": "folder"}`,
			errorContains: "max",
		},
		{
			name:          "missing type",
			payload:       `{"parent_id": "123e4567-e89b-12d3-a456-426614174000", "name": "folder"}`,
			errorContains: "Type",
		},
		{
			name:          "invalid type value",
			payload:       `{"parent_id": "123e4567-e89b-12d3-a456-426614174000", "name": "folder", "type": "invalid"}`,
			errorContains: "oneof",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := &MockService{}
			handler := NewHandler(mockService)
			c, w := test.SetupContext()

			c.Request.Body = io.NopCloser(bytes.NewBufferString(tt.payload))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("user_id", 1)

			// Act
			handler.CreateURL(c)

			// Assert
			require.Equal(t, http.StatusBadRequest, w.Code)

			var res map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Contains(t, res["error"], "invalid request body")
			assert.Contains(t, res["error"], tt.errorContains)
		})
	}
}
func TestHandler_CreateURL_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	requestBody := RequestBody{
		ParentID: "123e4567-e89b-12d3-a456-426614174001",
		Name:     "folder",
		Type:     "folder",
	}
	requestJSON, _ := json.Marshal(requestBody)

	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestJSON))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	mockService.On("CreateURL", &requestBody, 1).Return(assert.AnError)

	// Act
	handler.CreateURL(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, assert.AnError, c.Errors[0].Err)
	mockService.AssertExpectations(t)
}

func TestHandler_GetRootID_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	userID := 1
	expectedRootID := "123e4567-e89b-12d3-a456-426614174001"

	c.Set("user_id", userID)

	mockService.On("GetRootID", userID).Return(expectedRootID, nil)

	// Act
	handler.GetRootID(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)

	var response string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedRootID, response)
	mockService.AssertExpectations(t)
}
func TestHandler_GetRootID_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	userID := 1

	c.Set("user_id", userID)

	mockService.On("GetRootID", userID).Return("", assert.AnError)

	// Act
	handler.GetRootID(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, assert.AnError, c.Errors[0].Err)
	mockService.AssertExpectations(t)
}

func TestHandler_GetURL_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	userID := 1
	urlID := "123e4567-e89b-12d3-a456-426614174001"
	expectedResponse := &URLResponse{
		BaseURL: BaseURL{
			ID:        urlID,
			Name:      "node",
			Type:      "folder",
			URL:       nil,
			CreatedAt: time.Unix(0, 0).Format(time.RFC3339),
			UpdatedAt: time.Unix(0, 0).Format(time.RFC3339),
		},
		Parent: []BaseURL{
			{
				ID:        "parent-id",
				Name:      "parent",
				Type:      "folder",
				URL:       nil,
				CreatedAt: time.Unix(0, 0).Format(time.RFC3339),
				UpdatedAt: time.Unix(0, 0).Format(time.RFC3339),
			},
		},
		Children: []BaseURL{
			{
				ID:        "child1-id",
				Name:      "child1",
				Type:      "folder",
				URL:       nil,
				CreatedAt: time.Unix(0, 0).Format(time.RFC3339),
				UpdatedAt: time.Unix(0, 0).Format(time.RFC3339),
			},
			{
				ID:        "child2-id",
				Name:      "child2",
				Type:      "folder",
				URL:       nil,
				CreatedAt: time.Unix(0, 0).Format(time.RFC3339),
				UpdatedAt: time.Unix(0, 0).Format(time.RFC3339),
			},
		},
	}

	c.Params = gin.Params{{Key: "id", Value: urlID}}
	c.Set("user_id", userID)

	mockService.On("GetURL", urlID, userID).Return(expectedResponse, nil)

	// Act
	handler.GetURL(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)

	var response URLResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, *expectedResponse, response)
	mockService.AssertExpectations(t)
}
func TestHandler_GetURL_InvalidRequestURI(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Set("user_id", 1)

	// Act
	handler.GetURL(c)

	// Assert
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid request uri")
	assert.Contains(t, response["error"], "uuid")
}
func TestHandler_GetURL_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	urlID := "123e4567-e89b-12d3-a456-426614174001"

	c.Params = gin.Params{{Key: "id", Value: urlID}}
	c.Set("user_id", 1)

	mockService.On("GetURL", urlID, 1).Return(nil, assert.AnError)

	// Act
	handler.GetURL(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, assert.AnError, c.Errors[0].Err)
	mockService.AssertExpectations(t)
}

func TestHandler_ReplaceURL_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	urlID := "123e4567-e89b-12d3-a456-426614174001"
	requestBody := RequestBody{
		ParentID: "550e8400-e29b-41d4-a716-446655440001",
		Name:     "updated-folder",
		Type:     "folder",
	}
	requestJSON, _ := json.Marshal(requestBody)

	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestJSON))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: urlID}}
	c.Set("user_id", 1)

	mockService.On("ReplaceURL", urlID, &requestBody, 1).Return(nil)

	// Act
	handler.ReplaceURL(c)
	c.Writer.WriteHeaderNow()

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
func TestHandler_ReplaceURL_InvalidRequestURI(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Set("user_id", 1)

	// Act
	handler.ReplaceURL(c)

	// Assert
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid request uri")
	assert.Contains(t, response["error"], "uuid")
}
func TestHandler_ReplaceURL_InvalidRequestBody(t *testing.T) {
	tests := []struct {
		name          string
		payload       string
		errorContains string
	}{
		{
			name:          "missing parent_id",
			payload:       `{"name": "folder", "type": "folder"}`,
			errorContains: "ParentID",
		},
		{
			name:          "invalid parent_id format",
			payload:       `{"parent_id": "not-a-uuid", "name": "folder", "type": "folder"}`,
			errorContains: "uuid",
		},
		{
			name:          "missing name",
			payload:       `{"parent_id": "123e4567-e89b-12d3-a456-426614174000", "type": "folder"}`,
			errorContains: "Name",
		},
		{
			name:          "name too long",
			payload:       `{"parent_id": "123e4567-e89b-12d3-a456-426614174000", "name": "this name is definitely too long", "type": "folder"}`,
			errorContains: "max",
		},
		{
			name:          "missing type",
			payload:       `{"parent_id": "123e4567-e89b-12d3-a456-426614174000", "name": "folder"}`,
			errorContains: "Type",
		},
		{
			name:          "invalid type value",
			payload:       `{"parent_id": "123e4567-e89b-12d3-a456-426614174000", "name": "folder", "type": "invalid"}`,
			errorContains: "oneof",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := &MockService{}
			handler := NewHandler(mockService)
			c, w := test.SetupContext()

			urlID := "123e4567-e89b-12d3-a456-426614174001"
			c.Request.Body = io.NopCloser(bytes.NewBufferString(tt.payload))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: urlID}}
			c.Set("user_id", 1)

			// Act
			handler.ReplaceURL(c)

			// Assert
			require.Equal(t, http.StatusBadRequest, w.Code)

			var res map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Contains(t, res["error"], "invalid request body")
			assert.Contains(t, res["error"], tt.errorContains)
		})
	}
}
func TestHandler_ReplaceURL_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	urlID := "123e4567-e89b-12d3-a456-426614174001"
	requestBody := RequestBody{
		ParentID: "550e8400-e29b-41d4-a716-446655440001",
		Name:     "updated-folder",
		Type:     "folder",
	}
	requestJSON, _ := json.Marshal(requestBody)

	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestJSON))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: urlID}}
	c.Set("user_id", 1)

	mockService.On("ReplaceURL", urlID, &requestBody, 1).Return(assert.AnError)

	// Act
	handler.ReplaceURL(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, assert.AnError, c.Errors[0].Err)
	mockService.AssertExpectations(t)
}

func TestHandler_DeleteURL_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	urlID := "123e4567-e89b-12d3-a456-426614174001"

	c.Params = gin.Params{{Key: "id", Value: urlID}}
	c.Set("user_id", 1)

	mockService.On("DeleteURL", urlID, 1).Return(nil)

	// Act
	handler.DeleteURL(c)
	c.Writer.WriteHeaderNow()

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
func TestHandler_DeleteURL_InvalidRequestURI(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Set("user_id", 1)

	// Act
	handler.DeleteURL(c)

	// Assert
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid request uri")
	assert.Contains(t, response["error"], "uuid")
}
func TestHandler_DeleteURL_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	urlID := "123e4567-e89b-12d3-a456-426614174001"

	c.Params = gin.Params{{Key: "id", Value: urlID}}
	c.Set("user_id", 1)

	mockService.On("DeleteURL", urlID, 1).Return(assert.AnError)

	// Act
	handler.DeleteURL(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, assert.AnError, c.Errors[0].Err)
	mockService.AssertExpectations(t)
}
