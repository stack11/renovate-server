package conf

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"arhat.dev/pkg/confhelper"
	"golang.org/x/net/http/httpproxy"
)

type HTTPProxyConfig struct {
	HTTP    string `json:"http" yaml:"http"`
	HTTPS   string `json:"https" yaml:"https"`
	NoProxy string `json:"noProxy" yaml:"noProxy"`
	CGI     bool   `json:"cgi" yaml:"cgi"`
}

type HTTPClientConfig struct {
	Proxy *HTTPProxyConfig     `json:"proxy" yaml:"proxy"`
	TLS   confhelper.TLSConfig `json:"tls" yaml:"tls"`
}

func (c *HTTPClientConfig) NewClient() (*http.Client, error) {
	var proxy func(*http.Request) (*url.URL, error)
	if p := c.Proxy; p != nil {
		cfg := httpproxy.Config{
			HTTPProxy:  p.HTTP,
			HTTPSProxy: p.HTTPS,
			NoProxy:    p.NoProxy,
			CGI:        p.CGI,
		}

		pf := cfg.ProxyFunc()

		proxy = func(req *http.Request) (*url.URL, error) {
			return pf(req.URL)
		}
	} else {
		proxy = http.ProxyFromEnvironment
	}

	tlsConfig, err := c.TLS.GetTLSConfig(false)
	if err != nil {
		return nil, fmt.Errorf("failed to load tls config: %w", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: proxy,
			DialContext: (&net.Dialer{
				Timeout:       30 * time.Second,
				KeepAlive:     30 * time.Second,
				FallbackDelay: 300 * time.Millisecond,
			}).DialContext,

			ForceAttemptHTTP2:     tlsConfig != nil,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   0,
			ExpectContinueTimeout: 0,
			TLSClientConfig:       tlsConfig,

			DialTLSContext:         nil,
			DisableKeepAlives:      false,
			DisableCompression:     false,
			MaxConnsPerHost:        0,
			ResponseHeaderTimeout:  0,
			TLSNextProto:           nil,
			ProxyConnectHeader:     nil,
			MaxResponseHeaderBytes: 0,
			WriteBufferSize:        0,
			ReadBufferSize:         0,
		},
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}, nil
}

type APICommonConfig struct {
	BaseURL string `json:"baseURL" yaml:"baseURL"`

	Auth   APIAuthConfig    `json:"auth" yaml:"auth"`
	Client HTTPClientConfig `json:"client" yaml:"client"`
}

type WebhookConfig struct {
	Path   string `json:"path" yaml:"path"`
	Secret string `json:"secret" yaml:"secret"`
}

type GitConfig struct {
	User  string `json:"user" yaml:"user"`
	Email string `json:"email" yaml:"email"`
}

type GitHubConfig struct {
	API struct {
		APICommonConfig `json:",inline" yaml:",inline"`

		UploadURL string `json:"uploadURL" yaml:"uploadURL"`
	} `json:"api" yaml:"api"`

	Git      GitConfig       `json:"git" yaml:"git"`
	Webhook  WebhookConfig   `json:"webhook" yaml:"webhook"`
	Projects []ProjectConfig `json:"projects" yaml:"projects"`
}

type GitLabConfig struct {
	API struct {
		APICommonConfig `json:",inline" yaml:",inline"`
	} `json:"api" yaml:"api"`

	Git      GitConfig       `json:"git" yaml:"git"`
	Webhook  WebhookConfig   `json:"webhook" yaml:"webhook"`
	Projects []ProjectConfig `json:"projects" yaml:"projects"`
}
