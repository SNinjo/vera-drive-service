package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"vera-identity-service/internal/app"
	"vera-identity-service/internal/middleware"
	"vera-identity-service/internal/url"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var a *app.App

func TestMain(m *testing.M) {
	// Setup
	gin.SetMode(gin.TestMode)

	dbURL, closeDB, err := SetupPostgresql()
	if err != nil {
		log.Fatal(err)
	}

	identityService := SetupIdentityService("mock-token-secret")

	envs := map[string]string{
		"PORT":                 "8082",
		"DATABASE_URL":         dbURL,
		"IDENTITY_SERVICE_URL": identityService.URL,
		"ALLOWED_ORIGIN":       "http://mock-origin-1, http://mock-origin-2",
	}
	for key, value := range envs {
		err = os.Setenv(key, value)
		if err != nil {
			log.Fatal(err)
		}
	}

	a, err = app.InitApp()
	if err != nil {
		log.Fatal(err)
	}

	err = a.DB.AutoMigrate(&url.URLNode{})
	if err != nil {
		log.Fatal(err)
	}

	// Run
	code := m.Run()

	// Teardown
	a.Close()
	identityService.Close()
	closeDB()

	os.Exit(code)
}

func createTestRequest(method, path string, body interface{}, token string) (*http.Request, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return req, nil
}

func TestAPI_GetRootID_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	userID := 1
	rootID := "123e4567-e89b-12d3-a456-426614174001"
	root := url.URLNode{
		ID:     rootID,
		UserID: userID,
		Name:   "name",
		Type:   "folder",
		URL:    nil,
	}
	err = a.DB.Create(&root).Error
	require.NoError(t, err)

	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		middleware.UserClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: strconv.Itoa(userID),
			},
		},
	).SignedString([]byte("mock-token-secret"))
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("GET", "/urls/root-id", nil, token)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, fmt.Sprintf("\"%s\"", rootID), w.Body.String())
}
func TestAPI_GetRootID_NoRoot(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	userID := 1
	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		middleware.UserClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: strconv.Itoa(userID),
			},
		},
	).SignedString([]byte("mock-token-secret"))
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("GET", "/urls/root-id", nil, token)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Regexp(t, `^"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"$`, w.Body.String())
}

func TestAPI_GetURL_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	userID := 1
	nodeID := uuid.New().String()
	parent := url.URLNode{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      "name",
		Type:      "folder",
		URL:       nil,
		CreatedAt: time.Unix(1, 0),
		UpdatedAt: time.Unix(1, 0),
	}
	node := url.URLNode{
		ID:        nodeID,
		UserID:    userID,
		ParentID:  &parent.ID,
		Name:      "name",
		Type:      "folder",
		URL:       nil,
		CreatedAt: time.Unix(2, 0),
		UpdatedAt: time.Unix(2, 0),
	}
	child := url.URLNode{
		ID:        uuid.New().String(),
		UserID:    userID,
		ParentID:  &node.ID,
		Name:      "name",
		Type:      "folder",
		URL:       nil,
		CreatedAt: time.Unix(3, 0),
		UpdatedAt: time.Unix(3, 0),
	}
	err = a.DB.Create(&parent).Error
	require.NoError(t, err)
	err = a.DB.Create(&node).Error
	require.NoError(t, err)
	err = a.DB.Create(&child).Error
	require.NoError(t, err)

	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		middleware.UserClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: strconv.Itoa(userID),
			},
		},
	).SignedString([]byte("mock-token-secret"))
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("GET", "/urls/"+nodeID, nil, token)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var resp url.URLResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, node.ID, resp.ID)
	assert.Equal(t, node.Name, resp.Name)
	assert.Equal(t, node.Type, resp.Type)
	assert.Equal(t, node.URL, resp.URL)
	assert.Equal(t, node.CreatedAt.Format(time.RFC3339), resp.CreatedAt)
	assert.Equal(t, node.UpdatedAt.Format(time.RFC3339), resp.UpdatedAt)

	require.Len(t, resp.Parent, 1)
	assert.Equal(t, parent.ID, resp.Parent[0].ID)
	assert.Equal(t, parent.Name, resp.Parent[0].Name)
	assert.Equal(t, parent.Type, resp.Parent[0].Type)
	assert.Equal(t, parent.URL, resp.Parent[0].URL)
	assert.Equal(t, parent.CreatedAt.Format(time.RFC3339), resp.Parent[0].CreatedAt)
	assert.Equal(t, parent.UpdatedAt.Format(time.RFC3339), resp.Parent[0].UpdatedAt)

	require.Len(t, resp.Children, 1)
	assert.Equal(t, child.ID, resp.Children[0].ID)
	assert.Equal(t, child.Name, resp.Children[0].Name)
	assert.Equal(t, child.Type, resp.Children[0].Type)
	assert.Equal(t, child.URL, resp.Children[0].URL)
	assert.Equal(t, child.CreatedAt.Format(time.RFC3339), resp.Children[0].CreatedAt)
	assert.Equal(t, child.UpdatedAt.Format(time.RFC3339), resp.Children[0].UpdatedAt)
}

func TestAPI_CreateURL_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	userID := 1
	parentID := uuid.New().String()
	parent := url.URLNode{
		ID:     parentID,
		UserID: userID,
		Name:   "name",
		Type:   "folder",
		URL:    nil,
	}
	err = a.DB.Create(&parent).Error
	require.NoError(t, err)

	requestBody := url.RequestBody{
		ParentID: parentID,
		Name:     "new folder",
		Type:     "folder",
		URL:      nil,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(userID),
		},
	}).SignedString([]byte("mock-token-secret"))
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("POST", "/urls", requestBody, token)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	req, err = createTestRequest("GET", "/urls/"+parentID, nil, token)
	require.NoError(t, err)
	w = httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp url.URLResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Children, 1)
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, resp.Children[0].ID)
	assert.Equal(t, "new folder", resp.Children[0].Name)
	assert.Equal(t, "folder", resp.Children[0].Type)
	assert.Nil(t, resp.Children[0].URL)
	createdAt, err := time.Parse(time.RFC3339, resp.Children[0].CreatedAt)
	require.NoError(t, err)
	updatedAt, err := time.Parse(time.RFC3339, resp.Children[0].UpdatedAt)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), createdAt, time.Second)
	assert.WithinDuration(t, time.Now(), updatedAt, time.Second)
}
func TestAPI_CreateURL_WithURL(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	userID := 1
	parentID := uuid.New().String()
	parent := url.URLNode{
		ID:     parentID,
		UserID: userID,
		Name:   "name",
		Type:   "folder",
		URL:    nil,
	}
	err = a.DB.Create(&parent).Error
	require.NoError(t, err)

	requestBody := url.RequestBody{
		ParentID: parentID,
		Name:     "new url",
		Type:     "url",
		URL:      StringPtr("https://example.com"),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(userID),
		},
	}).SignedString([]byte("mock-token-secret"))
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("POST", "/urls", requestBody, token)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	req, err = createTestRequest("GET", "/urls/"+parentID, nil, token)
	require.NoError(t, err)
	w = httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp url.URLResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp.Children, 1)
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, resp.Children[0].ID)
	assert.Equal(t, "new url", resp.Children[0].Name)
	assert.Equal(t, "url", resp.Children[0].Type)
	assert.Equal(t, "https://example.com", *resp.Children[0].URL)
	createdAt, err := time.Parse(time.RFC3339, resp.Children[0].CreatedAt)
	require.NoError(t, err)
	updatedAt, err := time.Parse(time.RFC3339, resp.Children[0].UpdatedAt)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), createdAt, time.Second)
	assert.WithinDuration(t, time.Now(), updatedAt, time.Second)
}

func TestAPI_ReplaceURL_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	userID := 1
	nodeID := uuid.New().String()
	parentID := uuid.New().String()
	parent := url.URLNode{
		ID:     parentID,
		UserID: userID,
		Name:   "name",
		Type:   "folder",
		URL:    nil,
	}
	node := url.URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "name",
		Type:   "folder",
		URL:    nil,
	}
	err = a.DB.Create(&parent).Error
	require.NoError(t, err)
	err = a.DB.Create(&node).Error
	require.NoError(t, err)

	requestBody := url.RequestBody{
		ParentID: parentID,
		Name:     "new url",
		Type:     "url",
		URL:      StringPtr("https://example.com"),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(userID),
		},
	}).SignedString([]byte("mock-token-secret"))
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("PUT", "/urls/"+nodeID, requestBody, token)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)

	req, err = createTestRequest("GET", "/urls/"+nodeID, nil, token)
	require.NoError(t, err)
	w = httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp url.URLResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, nodeID, resp.ID)
	assert.Equal(t, "new url", resp.Name)
	assert.Equal(t, "url", resp.Type)
	assert.Equal(t, "https://example.com", *resp.URL)
	createdAt, err := time.Parse(time.RFC3339, resp.CreatedAt)
	require.NoError(t, err)
	updatedAt, err := time.Parse(time.RFC3339, resp.UpdatedAt)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), createdAt, time.Second)
	assert.WithinDuration(t, time.Now(), updatedAt, time.Second)

	require.Len(t, resp.Parent, 1)
	assert.Equal(t, parentID, resp.Parent[0].ID)
}

func TestAPI_DeleteURL_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	userID := 1
	nodeID := uuid.New().String()
	node := url.URLNode{
		ID:     nodeID,
		UserID: userID,
		Name:   "name",
		Type:   "folder",
		URL:    nil,
	}
	err = a.DB.Create(&node).Error
	require.NoError(t, err)

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(userID),
		},
	}).SignedString([]byte("mock-token-secret"))
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("DELETE", "/urls/"+nodeID, nil, token)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)

	req, err = createTestRequest("GET", "/urls/"+nodeID, nil, token)
	require.NoError(t, err)
	w = httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPI_AllURLs_Unauthorized(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{"POST", "/urls"},
		{"GET", "/urls/root-id"},
		{"GET", "/urls/123e4567-e89b-12d3-a456-426614174001"},
		{"PUT", "/urls/123e4567-e89b-12d3-a456-426614174001"},
		{"DELETE", "/urls/123e4567-e89b-12d3-a456-426614174001"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			// Arrange
			err := CleanupTables(a.DB)
			require.NoError(t, err)

			token, err := jwt.New(jwt.SigningMethodHS256).SignedString([]byte("invalid-token-secret"))
			require.NoError(t, err)

			// Act
			req, err := createTestRequest(tt.method, tt.path, nil, token)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			a.Router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "mock_error_code")
		})
	}
}
