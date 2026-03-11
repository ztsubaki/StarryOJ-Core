package structs

type User struct {
	Uid       uint64 `json:"uid" db:"uid"`
	Username  string `json:"username" db:"username"`
	Password  string `json:"-" db:"password"`
	Salt      string `json:"-" db:"salt"`
	LoginSalt string `json:"-" db:"login_salt"`
	Email     string `json:"email" db:"email"`
	Nickname  string `json:"nickname" db:"nickname"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}
