package github

import (
	"net/http"
	"strings"

	"github.com/google/go-github/v32/github"

	"arhat.dev/renovate-server/pkg/types"
	"arhat.dev/renovate-server/pkg/util"
)

func (m *Manager) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	payload, err := github.ValidatePayload(req, m.webhookSecret)
	if err != nil {
		http.Error(w, "invalid hmac signature", http.StatusBadRequest)
		return
	}

	e, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		http.Error(w, "invalid webhook payload", http.StatusBadRequest)
		return
	}

	repoAPIURL := func() string {
		switch evt := e.(type) {
		case *github.IssuesEvent:
			if evt.Action == nil || *evt.Action != "edited" {
				return ""
			}

			if evt.Changes == nil || evt.Changes.Body == nil || evt.Changes.Body.From == nil {
				return ""
			}

			if evt.GetIssue().GetUser().GetEmail() != m.gitEmail {
				return ""
			}

			oldBody := *evt.Changes.Body.From
			if util.ItemChecked(oldBody, evt.GetIssue().GetBody()) {
				return evt.GetRepo().GetURL()
			}

			return ""
		case *github.PullRequestEvent:
			if evt.Action == nil || *evt.Action != "edited" {
				return ""
			}

			if evt.Changes == nil || evt.Changes.Body == nil || evt.Changes.Body.From == nil {
				return ""
			}

			if evt.GetPullRequest().GetUser().GetEmail() != m.gitEmail {
				return ""
			}

			oldBody := *evt.Changes.Body.From
			if util.ItemChecked(oldBody, evt.GetPullRequest().GetBody()) {
				return evt.GetRepo().GetURL()
			}
			return ""
		case *github.PushEvent:
			return evt.GetRepo().GetURL()
		default:
			return ""
		}
	}()
	if repoAPIURL == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	repo := strings.TrimPrefix(repoAPIURL, m.apiURLPrefix+"/repos/")

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
		http.Error(w, "failed to execute renovate", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
