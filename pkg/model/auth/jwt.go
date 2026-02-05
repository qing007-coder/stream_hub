package auth

type Claims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`
}
