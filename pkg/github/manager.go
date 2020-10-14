package github

import (
	"context"
	"fmt"
	"net/http"

	"arhat.dev/pkg/log"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"

	"arhat.dev/renovate-server/pkg/conf"
	"arhat.dev/renovate-server/pkg/constant"
	"arhat.dev/renovate-server/pkg/types"
	"arhat.dev/renovate-server/pkg/util"
)

func NewManager(
	ctx context.Context,
	config *conf.PlatformConfig,
	scheduler types.Scheduler,
) (types.PlatformManager, error) {
	client, err := config.API.Client.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create http client")
	}

	var transport http.RoundTripper
	if o := config.API.OAuthToken; o != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: o,
		})

		transport = oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, client), ts).Transport
	} else {
		return nil, fmt.Errorf("no oauth token provided")
	}

	baseURL := config.API.BaseURL
	if baseURL == "" {
		baseURL = constant.DefaultGitHubAPIBaseURL
	}

	ghClient, err := github.NewEnterpriseClient(baseURL, "", &http.Client{
		Transport:     transport,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create github api client: %w", err)
	}

	dashboardTitles := make(map[string]string)
	for _, p := range config.Projects {
		dashboardTitles[p.Name] = p.DashboardIssueTitle
	}

	return &Manager{
		ctx: ctx,

		logger: log.Log.WithName("github").WithFields(
			log.String("path", config.Webhook.Path),
			log.String("api", config.API.BaseURL),
		),
		client:    ghClient,
		scheduler: scheduler,

		defaultDashboardTitle: config.DashboardIssueTitle,
		dashboardTitles:       dashboardTitles,

		apiURL:   baseURL,
		apiToken: config.API.OAuthToken,
		gitUser:  config.Git.User,
		gitEmail: config.Git.Email,

		webhookSecret: []byte(config.Webhook.Secret),
	}, nil
}

type Manager struct {
	ctx context.Context

	logger    log.Interface
	client    *github.Client
	scheduler types.Scheduler

	defaultDashboardTitle string
	dashboardTitles       map[string]string

	apiURL   string
	apiToken string
	gitUser  string
	gitEmail string

	webhookSecret []byte
}

func (m *Manager) getDashboardTitle(repo string) string {
	return util.GetOrDefault(m.dashboardTitles, repo, m.defaultDashboardTitle)
}
