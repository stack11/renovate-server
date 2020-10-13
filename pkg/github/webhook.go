package github

import (
	"net/http"
	"strings"

	"arhat.dev/pkg/log"
	"github.com/google/go-github/v32/github"

	"arhat.dev/renovate-server/pkg/types"
	"arhat.dev/renovate-server/pkg/util"
)

func (m *Manager) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := m.logger.WithFields()

	defer func() {
		err := recover()
		if err != nil {
			logger.E("recovered", log.Any("panic", err))
		}
	}()

	logger.D("event received")

	payload, err := github.ValidatePayload(req, m.webhookSecret)
	if err != nil {
		logger.I("signature invalid", log.Error(err))
		http.Error(w, "invalid hmac signature", http.StatusBadRequest)
		return
	}

	ev, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		logger.I("payload invalid", log.Error(err))
		http.Error(w, "invalid webhook payload", http.StatusBadRequest)
		return
	}

	repoAPIURL := func() string {
		switch evt := ev.(type) {
		case *github.IssuesEvent:
			logger.V("received issue event")
			if evt.Action == nil || *evt.Action != "edited" {
				return ""
			}

			logger.V("event is issue edited")

			if evt.Changes == nil || evt.Changes.Body == nil || evt.Changes.Body.From == nil {
				logger.V("issue body not changed")
				return ""
			}

			logger.V("issue body changed")

			if email := evt.GetIssue().GetUser().GetEmail(); email != m.gitEmail {
				logger.D("issue was not created by us",
					log.String("expect", m.gitEmail),
					log.String("actual", email),
				)
				return ""
			}

			logger.D("issue was created by us, checking checkbox state")

			oldBody := *evt.Changes.Body.From
			if util.ItemChecked(oldBody, evt.GetIssue().GetBody()) {
				return evt.GetRepo().GetURL()
			}
			return ""
		case *github.PullRequestEvent:
			logger.V("received pull request event")
			if evt.Action == nil || *evt.Action != "edited" {
				return ""
			}

			logger.V("event is pull request edited")

			if evt.Changes == nil || evt.Changes.Body == nil || evt.Changes.Body.From == nil {
				logger.V("pull request   unchanged")
				return ""
			}

			logger.V("pull request body changed")

			if evt.GetPullRequest().GetUser().GetEmail() != m.gitEmail {
				return ""
			}

			logger.D("pull request was created by us, checking checkbox state")

			oldBody := *evt.Changes.Body.From
			if util.ItemChecked(oldBody, evt.GetPullRequest().GetBody()) {
				return evt.GetRepo().GetURL()
			}
			return ""
		case *github.PushEvent:
			logger.V("received push event")
			return evt.GetRepo().GetURL()
		default:
			logger.V("received ignored event")
			return ""
		}
	}()
	if repoAPIURL == "" {
		logger.I("no execution triggered")
		w.WriteHeader(http.StatusOK)
		return
	}

	repo := strings.TrimPrefix(repoAPIURL, m.apiURLPrefix+"/repos/")

	logger = logger.WithFields(log.String("repo", repo), log.String("url", repoAPIURL))

	logger.I("executing renovate")

	// run renovate against this repo
	err = m.executor.Execute(types.ExecutionArgs{
		Platform: "github",
		APIURL:   m.apiURLPrefix,
		APIToken: m.apiToken,
		Repo:     repo,
		GitUser:  m.gitUser,
		GitEmail: m.gitEmail,
	})
	if err != nil {
		logger.I("failed to execute renovate", log.Error(err))
		http.Error(w, "failed to execute renovate", http.StatusInternalServerError)
		return
	}

	logger.I("executed renovate")
	w.WriteHeader(http.StatusOK)
}
