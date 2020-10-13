package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"

	"arhat.dev/renovate-server/pkg/conf"
	"arhat.dev/renovate-server/pkg/constant"
	"arhat.dev/renovate-server/pkg/types"
)

func NewManager(ctx context.Context, config *conf.GitHubConfig) (types.PlatformManager, error) {
	client, err := config.API.Client.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create http client")
	}

	var transport http.RoundTripper
	if b := config.API.Auth.Basic; b != nil {
		transport = &github.BasicAuthTransport{
			Username:  b.Username,
			Password:  b.Password,
			OTP:       b.OTP,
			Transport: client.Transport,
		}
	} else if o := config.API.Auth.OAuth; o != nil {
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: o.Token,
		})

		transport = oauth2.NewClient(ctx, ts).Transport
	} else {
		return nil, fmt.Errorf("no auth method provided")
	}

	baseURL, uploadURL := config.API.BaseURL, config.API.UploadURL
	if baseURL == "" {
		baseURL = constant.DefaultGitHubAPIBaseURL
		uploadURL = constant.DefaultGitHubAPIUploadURL
	}

	ghClient, err := github.NewEnterpriseClient(baseURL, uploadURL, &http.Client{
		Transport:     transport,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create github api client: %w", err)
	}

	return &Manager{
		ctx:    ctx,
		client: ghClient,
	}, nil
}

type Manager struct {
	ctx    context.Context
	client *github.Client
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "unimplemented", http.StatusNotImplemented)
}
