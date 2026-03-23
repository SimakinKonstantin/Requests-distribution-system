package model

// Client represents an end-user who submits appeals.
type Client struct {
	ID      int    `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	IsVIP   bool   `json:"isVip"`
}
