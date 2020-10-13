package gitlab

import (
	"context"
	"fmt"

	"arhat.dev/pkg/log"
	"github.com/xanzy/go-gitlab"

	"arhat.dev/renovate-server/pkg/conf"
	"arhat.dev/renovate-server/pkg/constant"
	"arhat.dev/renovate-server/pkg/types"
)

func NewManager(
	ctx context.Context,
	config *conf.GitLabConfig,
	executor types.Executor,
) (types.PlatformManager, error) {
	client, err := config.API.Client.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create http client")
	}

	baseURL := config.API.BaseURL
	if baseURL == "" {
		baseURL = constant.DefaultGitLabAPIBaseURL
	}

	var glClient *gitlab.Client
	//if b := config.API.Auth.Basic; b != nil {
	//	glClient, err = gitlab.NewBasicAuthClient(b.Username, b.Password,
	//		gitlab.WithBaseURL(baseURL),
	//		gitlab.WithHTTPClient(client),
	//	)
	//}

	if o := config.API.Auth.OAuth; o != nil {
		glClient, err = gitlab.NewOAuthClient(o.Token,
			gitlab.WithBaseURL(baseURL),
			gitlab.WithHTTPClient(client),
		)
	} else {
		return nil, fmt.Errorf("no oauth token provided")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %w", err)
	}

	return &Manager{
		ctx: ctx,

		logger: log.Log.WithName("gitlab").WithFields(
			log.String("path", config.Webhook.Path),
			log.String("api", config.API.BaseURL),
		),
		client:   glClient,
		executor: executor,

		apiURL:   baseURL,
		apiToken: config.API.Auth.OAuth.Token,
		gitUser:  config.Git.User,
		gitEmail: config.Git.Email,
	}, nil
}

type Manager struct {
	ctx context.Context

	logger   log.Interface
	client   *gitlab.Client
	executor types.Executor

	apiURL   string
	apiToken string
	gitUser  string
	gitEmail string
}
