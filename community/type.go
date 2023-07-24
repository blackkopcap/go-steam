package community

type LoginDetails struct {
	AccountName   string
	Password      string
	TwoFactorCode string
}

// responses

type getRSAKeyRes struct {
	Success      bool   `json:"success"`
	PublickeyMod string `json:"publickey_mod"`
	PublickeyExp string `json:"publickey_exp"`
	Timestamp    string `json:"timestamp"`
	TokenGid     string `json:"token_gid"`
}

type loginSession struct {
	Success           bool   `json:"success"`
	LoginComplete     bool   `json:"login_complete"`
	RequiresTwoFactor bool   `json:"requires_twofactor"`
	Message           string `json:"message"`
	RedirectURI       string `json:"redirect_uri"`
	OAuth             oAuth  `json:"transfer_parameters"`
}

type oAuth struct {
	ID               string `json:"-"`
	DeviceID         string `json:"-"`
	SteamID          string `json:"steamid"`
	Auth             string `json:"auth"`
	TokenSecure      string `json:"token_secure"`
	WebCookie        string `json:"webcookie"`
	SteamLogin       string `json:"-"`
	SteamLoginSecure string `json:"-"`
}
