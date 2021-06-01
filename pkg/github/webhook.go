package github

import (
	"net/http"

	"arhat.dev/pkg/log"
	"github.com/google/go-github/v35/github"

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

	repo := func() string {
		switch evt := ev.(type) {
		case *github.IssuesEvent:
			repo := evt.GetRepo().GetFullName()
			logger = logger.WithFields(log.String("repo", repo))
			logger.V("received issue event")

			expectedTitle := m.getDashboardTitle(repo)
			if expectedTitle == "" {
				// no dashboard issue title provided, we may assume any issue with any title can trigger
				// if they have checkbox (todo list)
			} else if actualTitle := evt.GetIssue().GetTitle(); expectedTitle != actualTitle {
				logger.D("issue event is not related to renovate dashboard issue",
					log.String("expected", expectedTitle),
					log.String("actual", actualTitle),
				)
				return ""
			}

			if evt.Action == nil {
				// unknown action, just trigger the execution
				return repo
			}

			switch *evt.Action {
			case "edited":
				logger.V("event is issue edited")
			case "deleted", "transferred", "closed", "reopened":
				// dashboard issue state changed, ensure open
				return repo
			default:
				return ""
			}

			if evt.Changes == nil || evt.Changes.Body == nil || evt.Changes.Body.From == nil {
				logger.V("issue body not changed")
				return ""
			}

			logger.D("issue body changed, checking issue checkbox state")
			oldBody := *evt.Changes.Body.From
			if util.ItemChecked(oldBody, evt.GetIssue().GetBody()) {
				return repo
			}
			return ""
		case *github.PullRequestEvent:
			repo := evt.GetRepo().GetFullName()
			logger = logger.WithFields(log.String("repo", repo))

			logger.V("received pull request event")
			if evt.Action == nil {
				return repo
			}

			switch *evt.Action {
			case "edited":
				logger.V("event is pull request edited")
			case "closed", "reopened":
				return repo
			default:
				return ""
			}

			if evt.Changes == nil || evt.Changes.Body == nil || evt.Changes.Body.From == nil {
				logger.V("pull request body unchanged")
				return ""
			}

			logger.V("pull request body changed")

			oldBody := *evt.Changes.Body.From
			if util.ItemChecked(oldBody, evt.GetPullRequest().GetBody()) {
				return repo
			}
			return ""
		case *github.PushEvent:
			repo := evt.GetRepo().GetFullName()
			logger = logger.WithFields(log.String("repo", repo))
			logger.V("received push event")

			return repo
		default:
			logger.V("ignored event")
			return ""
		}
	}()
	if repo == "" {
		logger.I("no execution triggered")
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, disabled := m.disabledRepos[repo]; disabled {
		logger.I("execution ignored")
		w.WriteHeader(http.StatusOK)
		return
	}

	logger.I("scheduling renovate execution")

	// run renovate against this repo
	err = m.scheduler.Schedule(m.ExecutionArgs(repo))
	if err != nil {
		logger.I("failed to schedule renovate execution", log.Error(err))
		http.Error(w, "failed to execute renovate", http.StatusInternalServerError)
		return
	}

	logger.I("scheduled renovate execution")
	w.WriteHeader(http.StatusOK)
}
