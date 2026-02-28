package model

// Client represents an end-user who submits appeals.
type Client struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}
