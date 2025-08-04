package url

import "time"

type RequestURI struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type RequestBody struct {
	ParentID string  `json:"parent_id" binding:"required,uuid"`
	Name     string  `json:"name" binding:"required,max=20"`
	Type     string  `json:"type" binding:"required,oneof=folder url"`
	URL      *string `json:"url"`
}

type BaseURL struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	URL       *string `json:"url"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type URLResponse struct {
	BaseURL
	Parent   []BaseURL `json:"parent"`
	Children []BaseURL `json:"children"`
}

func newBaseURL(node *URLNode) *BaseURL {
	return &BaseURL{
		ID:        node.ID,
		Name:      node.Name,
		Type:      node.Type,
		URL:       node.URL,
		CreatedAt: node.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: node.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
func newURLResponse(node *URLNode, parents []URLNode, children []URLNode) *URLResponse {
	newParents := make([]BaseURL, len(parents))
	for i, parent := range parents {
		newParents[i] = *newBaseURL(&parent)
	}
	newChildren := make([]BaseURL, len(children))
	for i, child := range children {
		newChildren[i] = *newBaseURL(&child)
	}

	return &URLResponse{
		BaseURL:  *newBaseURL(node),
		Parent:   newParents,
		Children: newChildren,
	}
}
