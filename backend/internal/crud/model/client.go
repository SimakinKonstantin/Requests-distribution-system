package model

type Client struct {
	ID      int    `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	IsVIP   bool   `json:"isVip"`
}
