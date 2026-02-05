package api

type LoginReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type SendVerifyCodeReq struct {
	Email string `json:"email"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}

type RegisterReq struct {
	Account          string `json:"account"`
	Password         string `json:"password"`
	Email            string `json:"email"`
	VerificationCode string `json:"verification_code"`
}

type UpdateUserProfileReq struct {
	Nickname      string `json:"nickname"`
	AvatarUrl     string `json:"avatar_url"`
	BackgroundUrl string `json:"background_url"`
	Signature     string `json:"signature"`
	Gender        int8   `json:"gender"`
}

type UpdatePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type GetUserProfileResp struct {
	Account        string `json:"account"`
	Email          string `json:"email"`
	Nickname       string `json:"nickname"`
	BackgroundUrl  string `json:"background_url"`
	AvatarUrl      string `json:"avatar_url"`
	Gender         int8   `json:"gender"`
	Signature      string `json:"signature"`
	FollowCount    int64  `json:"follow_count"`
	FollowerCount  int64  `json:"follower_count"`
	WorkCount      int64  `json:"work_count"`
	FavouriteCount int64  `json:"favourite_count"`
}
