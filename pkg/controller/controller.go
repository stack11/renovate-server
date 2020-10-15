package controller

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/queue"
	"github.com/robfig/cron/v3"

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
		exec, err = executor.NewKubernetesExecutor(ctx, config.Server.Executor.Kubernetes)
	default:
		return nil, fmt.Errorf("no executor provided")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	tlsConfig, err := config.Server.Webhook.TLS.GetTLSConfig(true)
	if err != nil {
		return nil, fmt.Errorf("failed to create tls config for webhook server: %w", err)
	}

	var cronJob *cron.Cron
	if config.Server.Scheduling.Cron != "" {
		location := time.UTC
		if config.Server.Scheduling.Timezone != "" {
			location, err = time.LoadLocation(config.Server.Scheduling.Timezone)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timezone: %w", err)
			}
		}

		cronJob = cron.New(
			cron.WithLocation(location),
			cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)),
			cron.WithParser(
				cron.NewParser(
					cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
				),
			),
		)
	}

	ctrl := &Controller{
		ctx: ctx,

		logger:     log.Log.WithName("controller"),
		listenAddr: config.Server.Webhook.Listen,
		managers:   make(map[string]types.PlatformManager),
		tlsConfig:  tlsConfig,

		delay:    config.Server.Scheduling.Delay,
		executor: exec,
		tq:       queue.NewTimeoutQueue(),

		cronTab: config.Server.Scheduling.Cron,
		cronJob: cronJob,
	}

	for i, gh := range config.GitHub {
		mgr, err2 := github.NewManager(ctx, &config.GitHub[i], ctrl)
		if err2 != nil {
			return nil, fmt.Errorf("failed to create github manager, index %d: %w", i, err2)
		}
		ctrl.managers[gh.Webhook.Path] = mgr
	}

	for i, gh := range config.GitLab {
		mgr, err2 := gitlab.NewManager(ctx, &config.GitLab[i], ctrl)
		if err2 != nil {
			return nil, fmt.Errorf("failed to create gitlab manager, index %d: %w", i, err2)
		}
		ctrl.managers[gh.Webhook.Path] = mgr
	}

	return ctrl, nil
}

type Controller struct {
	ctx context.Context

	logger     log.Interface
	listenAddr string
	managers   map[string]types.PlatformManager
	tlsConfig  *tls.Config

	delay    time.Duration
	executor types.Executor
	tq       *queue.TimeoutQueue

	cronTab string
	cronJob *cron.Cron
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

	c.tq.Start(c.ctx.Done())
	go func() {
		ch := c.tq.TakeCh()
		for d := range ch {
			args := d.Data.(types.ExecutionArgs)
			c.logger.I("executing renovate")
			err2 := c.executor.Execute(args)
			if err2 != nil {
				c.logger.I("failed to execute renovate for repo, rescheduling",
					log.Strings("repos", args.Repos),
					log.String("endpoint", args.APIURL),
					log.Error(err),
				)
				_ = c.Schedule(args)
			} else {
				c.logger.I("finished renovate execution")
			}
		}
	}()

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

	if c.cronJob != nil {
		_, err = c.cronJob.AddFunc(c.cronTab, func() {
			c.logger.I("working on cron job")

			c.CheckAllRepos()

			c.logger.I("cron job finished")
		})
		if err != nil {
			return err
		}

		c.cronJob.Start()
	}

	return nil
}

func (c *Controller) CheckAllRepos() {
	wg := new(sync.WaitGroup)
	for k := range c.managers {
		wg.Add(1)

		go func(key string) {
			defer wg.Done()

			logger := c.logger.WithFields(
				log.String("job", "cron"),
				log.String("endpoint", key),
			)
			mgr := c.managers[key]
			repos, err2 := mgr.ListRepos()
			if err2 != nil {
				logger.I("failed to list repos", log.Error(err2))
				return
			}

			args := mgr.ExecutionArgs(repos...)
			err2 = c.executor.Execute(args)
			if err2 != nil {
				logger.I("failed to execute all repos check job, schedule as normal job", log.Error(err2))
				err2 = c.Schedule(args)
				if err2 != nil {
					logger.I("failed to schedule all repos check job as normal job", log.Error(err2))
				}
				return
			}
		}(k)
	}

	wg.Wait()
}

func (c *Controller) Schedule(args types.ExecutionArgs) error {
	key := args.APIURL + args.APIToken
	repos := args.Repos

	oldArgs, removed := c.tq.Remove(key)
	if removed {
		oldRepos := oldArgs.(types.ExecutionArgs).Repos

		for _, oldR := range oldRepos {
			found := false
			for _, r := range args.Repos {
				if oldR == r {
					found = true
				}
			}

			if !found {
				repos = append(repos, oldR)
			}
		}
	}

	args.Repos = repos
	return c.tq.OfferWithDelay(key, args, c.delay)
}
