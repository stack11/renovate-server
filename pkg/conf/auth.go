package conf

type OAuthConfig struct {
	Token string `json:"token" yaml:"token"`
}

type APIAuthConfig struct {
	OAuth *OAuthConfig `json:"oauth" yaml:"oauth"`
}
