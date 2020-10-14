package github

import (
	"net/http"

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

	repo := func() string {
		switch evt := ev.(type) {
		case *github.IssuesEvent:
			repo := evt.GetRepo().GetFullName()
			logger = logger.WithFields(log.String("repo", repo))
			logger.V("received issue event")

			expectedTitle := m.getDashboardTitle(repo)
			if actualTitle := evt.GetIssue().GetTitle(); expectedTitle != actualTitle {
				logger.D("issue event is not related to renovate dashboard issue",
					log.String("expected", expectedTitle),
					log.String("actual", actualTitle),
				)
				return ""
			}

			if evt.Action == nil {
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

	logger.I("executing renovate")

	// run renovate against this repo
	err = m.executor.Execute(types.ExecutionArgs{
		Platform: "github",
		APIURL:   m.apiURL,
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
