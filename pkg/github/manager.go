package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

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
	var (
		err error

		disabledRepoNameMatch *regexp.Regexp
	)
	if config.DisabledRepoNameMatch != "" {
		disabledRepoNameMatch, err = regexp.Compile(config.DisabledRepoNameMatch)
		if err != nil {
			return nil, fmt.Errorf("failed to compile disabled repo match: %w", err)
		}
	}

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

	ghClient.BaseURL, _ = url.Parse(baseURL)

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

		logger: log.Log.WithName("github").WithFields(
			log.String("path", config.Webhook.Path),
			log.String("api", config.API.BaseURL),
		),
		client:    ghClient,
		scheduler: scheduler,

		disabledRepoNameMatch: disabledRepoNameMatch,
		defaultDashboardTitle: config.DashboardIssueTitle,
		dashboardTitles:       dashboardTitles,
		disabledRepos:         disabledRepos,

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

	disabledRepoNameMatch *regexp.Regexp
	defaultDashboardTitle string
	dashboardTitles       map[string]string
	disabledRepos         map[string]struct{}

	apiURL   string
	apiToken string
	gitUser  string
	gitEmail string

	webhookSecret []byte
}

func (m *Manager) getDashboardTitle(repo string) string {
	return util.GetOrDefault(m.dashboardTitles, repo, m.defaultDashboardTitle)
}

func (m *Manager) ListRepos() ([]string, error) {
	repos, _, err := m.client.Repositories.List(m.ctx, "", &github.RepositoryListOptions{
		Visibility:  "",
		Affiliation: "",
		Type:        "",
		Sort:        "",
		Direction:   "",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 0,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list all repos: %w", err)
	}

	var ret []string
	for _, repo := range repos {
		name := repo.GetFullName()
		if m.disabledRepoNameMatch != nil {
			if m.disabledRepoNameMatch.Match([]byte(name)) {
				continue
			}
		}

		if _, disabled := m.disabledRepos[name]; disabled {
			continue
		}

		ret = append(ret, name)
	}

	return ret, nil
}

func (m *Manager) ExecutionArgs(repos ...string) types.ExecutionArgs {
	return types.ExecutionArgs{
		Platform: "github",
		APIURL:   m.apiURL,
		APIToken: m.apiToken,
		Repos:    repos,
		GitUser:  m.gitUser,
		GitEmail: m.gitEmail,
	}
}
