// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type ImageUploadResponse struct {
	Base64   string `json:"base64"`
	MimeType string `json:"mimeType"`
}

type MetaInfo struct {
	IPAddress    *string `json:"ipAddress,omitempty"`
	UserAgent    *string `json:"userAgent,omitempty"`
	CreatedEpoch *int    `json:"createdEpoch,omitempty"`
}

type Mutation struct {
}

type NewTodoParams struct {
	Text   string `json:"text"`
	UserID string `json:"userId"`
}

type Query struct {
}
