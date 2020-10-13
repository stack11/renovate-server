package controller

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"

	"arhat.dev/renovate-server/pkg/conf"
	"arhat.dev/renovate-server/pkg/executor"
	"arhat.dev/renovate-server/pkg/github"
	"arhat.dev/renovate-server/pkg/gitlab"
	"arhat.dev/renovate-server/pkg/types"
)

func NewController(ctx context.Context, config *conf.Config) (*Controller, error) {
	var (
		exec types.Executor
		err  error
	)
	switch {
	case config.Server.Executor.Kubernetes != nil:
		exec, err = executor.NewKubernetesExecutor(config.Server.Executor.Kubernetes)
	default:
		return nil, fmt.Errorf("no executor provided")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	managers := make(map[string]types.PlatformManager)
	for i, gh := range config.GitHub {
		mgr, err2 := github.NewManager(ctx, &config.GitHub[i], exec)
		if err2 != nil {
			return nil, fmt.Errorf("failed to create github manager, index %d: %w", i, err2)
		}
		managers[gh.Webhook.Path] = mgr
	}

	for i, gh := range config.GitLab {
		mgr, err2 := gitlab.NewManager(ctx, &config.GitLab[i], exec)
		if err2 != nil {
			return nil, fmt.Errorf("failed to create gitlab manager, index %d: %w", i, err2)
		}
		managers[gh.Webhook.Path] = mgr
	}

	tlsConfig, err := config.Server.Webhook.TLS.GetTLSConfig(true)
	if err != nil {
		return nil, fmt.Errorf("failed to create tls config for webhook server: %w", err)
	}

	return &Controller{
		ctx:        ctx,
		listenAddr: config.Server.Webhook.Listen,
		managers:   managers,
		tlsConfig:  tlsConfig,
	}, nil
}

type Controller struct {
	ctx        context.Context
	listenAddr string
	managers   map[string]types.PlatformManager
	tlsConfig  *tls.Config
}

func (c *Controller) Start() error {
	mux := http.NewServeMux()
	for path := range c.managers {
		mux.Handle(path, c.managers[path])
	}

	srv := &http.Server{
		Handler:   mux,
		TLSConfig: c.tlsConfig,
		BaseContext: func(listener net.Listener) context.Context {
			return c.ctx
		},
	}

	l, err := net.Listen("tcp", c.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen tcp for webhook server: %w", err)
	}

	if c.tlsConfig != nil {
		l = tls.NewListener(l, c.tlsConfig)
	}

	go func() {
		err2 := srv.Serve(l)
		if err2 != nil && errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("failed to serve webhook server: %w", err))
		}
	}()

	go func() {
		defer func() {
			_ = srv.Close()
		}()

		// nolint:gosimple
		select {
		case <-c.ctx.Done():
		}
	}()

	return nil
}
