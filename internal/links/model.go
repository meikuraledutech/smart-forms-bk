package links

// PublicForm represents the public view of a form
type PublicForm struct {
	ID                 string                 `json:"id"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description"`
	AcceptingResponses bool                   `json:"accepting_responses"`
	Flow               map[string]interface{} `json:"flow"`
}

// PublishRequest represents the request to publish a form
type PublishRequest struct {
	CustomSlug string `json:"custom_slug,omitempty"`
}

// PublishResponse represents the response after publishing
type PublishResponse struct {
	AutoSlug   string  `json:"auto_slug"`
	CustomSlug *string `json:"custom_slug,omitempty"`
	AutoURL    string  `json:"auto_url"`
	CustomURL  *string `json:"custom_url,omitempty"`
}

// ToggleRequest represents the request to toggle accepting responses
type ToggleRequest struct {
	Accepting bool `json:"accepting"`
}
