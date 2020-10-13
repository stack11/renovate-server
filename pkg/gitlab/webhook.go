package gitlab

import (
	"io/ioutil"
	"net/http"

	"github.com/xanzy/go-gitlab"

	"arhat.dev/renovate-server/pkg/types"
	"arhat.dev/renovate-server/pkg/util"
)

func (m *Manager) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "failed to read webhook body", http.StatusBadRequest)
		return
	}

	ev, err := gitlab.ParseHook(gitlab.HookEventType(req), payload)
	if err != nil {
		http.Error(w, "invalid webhook body", http.StatusBadRequest)
		return
	}

	repo := func() string {
		switch evt := ev.(type) {
		case *gitlab.IssueEvent:
			if evt.User == nil || evt.User.Email != m.gitEmail {
				return ""
			}

			if util.CountCheckedItems(evt.ObjectAttributes.Description) > 0 {
				return evt.Project.PathWithNamespace
			}

			return ""
		case *gitlab.MergeEvent:
			if evt.User == nil || evt.User.Email != m.gitEmail {
				return ""
			}

			if util.CountCheckedItems(evt.ObjectAttributes.Description) > 0 {
				return evt.Project.PathWithNamespace
			}

			return ""
		case *gitlab.PushEvent:
			if evt.UserEmail == m.gitEmail {
				return ""
			}

			return evt.Project.PathWithNamespace
		default:
			return ""
		}
	}()
	if repo == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = m.executor.Execute(types.ExecutionArgs{
		Platform: "gitlab",
		APIURL:   m.apiURL,
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
