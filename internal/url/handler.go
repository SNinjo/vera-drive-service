package url

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateURL(c *gin.Context) {
	var body RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body | " + err.Error()})
		return
	}

	userID := c.GetInt("user_id")
	if err := h.service.CreateURL(&body, userID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetRootID(c *gin.Context) {
	userID := c.GetInt("user_id")
	fmt.Println("GetRootID", userID)

	rootID, err := h.service.GetRootID(userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, rootID)
}

func (h *Handler) GetURL(c *gin.Context) {
	uri := &RequestURI{}
	if err := c.ShouldBindUri(uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request uri | " + err.Error()})
		return
	}

	userID := c.GetInt("user_id")
	response, err := h.service.GetURL(uri.ID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) ReplaceURL(c *gin.Context) {
	uri := &RequestURI{}
	if err := c.ShouldBindUri(uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request uri | " + err.Error()})
		return
	}

	var body RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body | " + err.Error()})
		return
	}

	userID := c.GetInt("user_id")
	if err := h.service.ReplaceURL(uri.ID, &body, userID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) DeleteURL(c *gin.Context) {
	uri := &RequestURI{}
	if err := c.ShouldBindUri(uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request uri | " + err.Error()})
		return
	}

	userID := c.GetInt("user_id")
	if err := h.service.DeleteURL(uri.ID, userID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
