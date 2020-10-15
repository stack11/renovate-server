package gitlab

import (
	"context"
	"fmt"

	"arhat.dev/pkg/log"
	"github.com/xanzy/go-gitlab"

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

	baseURL := config.API.BaseURL
	if baseURL == "" {
		baseURL = constant.DefaultGitLabAPIBaseURL
	}

	var glClient *gitlab.Client
	if o := config.API.OAuthToken; o != "" {
		glClient, err = gitlab.NewOAuthClient(o,
			gitlab.WithBaseURL(baseURL),
			gitlab.WithHTTPClient(client),
		)
	} else {
		return nil, fmt.Errorf("no oauth token provided")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client: %w", err)
	}

	dashboardTitles := make(map[string]string)
	disabledRepos := make(map[string]struct{})
	for _, p := range config.Projects {
		dashboardTitles[p.Name] = p.DashboardIssueTitle
		if p.Disabled {
			disabledRepos[p.Name] = struct{}{}
		}
	}

	return &Manager{
		ctx: ctx,

		logger: log.Log.WithName("gitlab").WithFields(
			log.String("path", config.Webhook.Path),
			log.String("api", config.API.BaseURL),
		),
		client:    glClient,
		scheduler: scheduler,

		defaultDashboardTitle: config.DashboardIssueTitle,
		dashboardTitles:       dashboardTitles,
		disabledRepos:         disabledRepos,

		apiURL:   baseURL,
		apiToken: config.API.OAuthToken,
		gitUser:  config.Git.User,
		gitEmail: config.Git.Email,
	}, nil
}

type Manager struct {
	ctx context.Context

	logger    log.Interface
	client    *gitlab.Client
	scheduler types.Scheduler

	defaultDashboardTitle string
	dashboardTitles       map[string]string
	disabledRepos         map[string]struct{}

	apiURL   string
	apiToken string
	gitUser  string
	gitEmail string
}

func (m *Manager) getDashboardTitle(repo string) string {
	return util.GetOrDefault(m.dashboardTitles, repo, m.defaultDashboardTitle)
}

func (m *Manager) ListRepos() ([]string, error) {
	falseP := false
	trueP := true

	repos, _, err := m.client.Projects.ListProjects(&gitlab.ListProjectsOptions{
		Archived: &falseP,
		Simple:   &trueP,
	}, gitlab.WithContext(m.ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list all repos: %w", err)
	}

	var ret []string
	for _, repo := range repos {
		name := repo.PathWithNamespace
		if _, disabled := m.disabledRepos[name]; disabled {
			continue
		}

		ret = append(ret, name)
	}

	return ret, nil
}

func (m *Manager) ExecutionArgs(repos ...string) types.ExecutionArgs {
	return types.ExecutionArgs{
		Platform: "gitlab",
		APIURL:   m.apiURL,
		APIToken: m.apiToken,
		Repos:    repos,
		GitUser:  m.gitUser,
		GitEmail: m.gitEmail,
	}
}
