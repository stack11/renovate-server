package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"github.com/xanzy/go-gitlab"

	"arhat.dev/renovate-server/pkg/conf"
	"arhat.dev/renovate-server/pkg/constant"
	"arhat.dev/renovate-server/pkg/types"
)

func NewManager(ctx context.Context, config *conf.GitLabConfig) (types.PlatformManager, error) {
	client, err := config.API.Client.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create http client")
	}

	baseURL := config.API.BaseURL
	if baseURL == "" {
		baseURL = constant.DefaultGitLabAPIBaseURL
	}

	var glClient *gitlab.Client
	if b := config.API.Auth.Basic; b != nil {
		glClient, err = gitlab.NewBasicAuthClient(b.Username, b.Password,
			gitlab.WithBaseURL(baseURL),
			gitlab.WithHTTPClient(client),
		)
	} else if o := config.API.Auth.OAuth; o != nil {
		glClient, err = gitlab.NewOAuthClient(o.Token,
			gitlab.WithBaseURL(baseURL),
			gitlab.WithHTTPClient(client),
		)
	} else {
		return nil, fmt.Errorf("no auth method provided")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %w", err)
	}

	return &Manager{
		ctx:    ctx,
		client: glClient,
	}, nil
}

type Manager struct {
	ctx    context.Context
	client *gitlab.Client
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "unimplemented", http.StatusNotImplemented)
}
