package types

type SMTPSettings struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	RecipientEmail string `json:"recipient_email"`
}
