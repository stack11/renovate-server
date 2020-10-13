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
)

func NewManager(
	ctx context.Context,
	config *conf.GitHubConfig,
	executor types.Executor,
) (types.PlatformManager, error) {
	client, err := config.API.Client.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create http client")
	}

	var transport http.RoundTripper

	// currently there is no basic auth support in renovate
	//if b := config.API.Auth.Basic; b != nil {
	//	transport = &github.BasicAuthTransport{
	//		Username:  b.Username,
	//		Password:  b.Password,
	//		OTP:       b.OTP,
	//		Transport: client.Transport,
	//	}
	//}

	if o := config.API.Auth.OAuth; o != nil {
		ts := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: o.Token,
		})

		transport = oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, client), ts).Transport
	} else {
		return nil, fmt.Errorf("no oauth token provided")
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

	apiURLPrefix := baseURL
	for apiURLPrefix[len(apiURLPrefix)-1] == '/' {
		apiURLPrefix = apiURLPrefix[:len(apiURLPrefix)-1]
	}

	return &Manager{
		ctx: ctx,

		logger: log.Log.WithName("github").WithFields(
			log.String("path", config.Webhook.Path),
			log.String("api", config.API.BaseURL),
		),
		client:   ghClient,
		executor: executor,

		apiURLPrefix: apiURLPrefix,
		apiToken:     config.API.Auth.OAuth.Token,
		gitUser:      config.Git.User,
		gitEmail:     config.Git.Email,

		webhookSecret: []byte(config.Webhook.Secret),
	}, nil
}

type Manager struct {
	ctx context.Context

	logger   log.Interface
	client   *github.Client
	executor types.Executor

	apiURLPrefix string
	apiToken     string
	gitUser      string
	gitEmail     string

	webhookSecret []byte
}
