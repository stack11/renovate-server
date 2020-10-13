package conf

type BasicAuthConfig struct {
	Username string `json:"username" yaml:"password"`
	Password string `json:"password" yaml:"password"`

	OTP string `json:"otp" yaml:"otp"`
}

type OAuthConfig struct {
	Token string `json:"token" yaml:"token"`
}

type APIAuthConfig struct {
	OAuth *OAuthConfig     `json:"oauth" yaml:"oauth"`
	Basic *BasicAuthConfig `json:"basic" yaml:"basic"`
}
