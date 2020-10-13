package gitlab

import (
	"io/ioutil"
	"net/http"

	"arhat.dev/pkg/log"
	"github.com/xanzy/go-gitlab"

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

	logger.D("received event")

	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.I("failed to read event payload", log.Error(err))
		http.Error(w, "failed to read event payload", http.StatusBadRequest)
		return
	}

	ev, err := gitlab.ParseHook(gitlab.HookEventType(req), payload)
	if err != nil {
		logger.I("event payload invalid", log.Error(err))
		http.Error(w, "invalid event payload", http.StatusBadRequest)
		return
	}

	repo := func() string {
		switch evt := ev.(type) {
		case *gitlab.IssueEvent:
			logger.V("received issue event")

			if evt.User == nil || evt.User.Email != m.gitEmail {
				logger.D("issue was not created by us",
					log.String("expect", m.gitEmail),
					log.String("actual", evt.User.Email),
				)
				return ""
			}

			logger.D("issue was created by us, checking checkbox state")

			if util.CountCheckedItems(evt.ObjectAttributes.Description) > 0 {
				return evt.Project.PathWithNamespace
			}

			return ""
		case *gitlab.MergeEvent:
			logger.V("received merge request event")

			if evt.User == nil || evt.User.Email != m.gitEmail {
				logger.D("pull request was not created by us",
					log.String("expect", m.gitEmail),
					log.String("actual", evt.User.Email),
				)
				return ""
			}

			logger.D("pull request was created by us, checking checkbox state")

			if util.CountCheckedItems(evt.ObjectAttributes.Description) > 0 {
				return evt.Project.PathWithNamespace
			}

			return ""
		case *gitlab.PushEvent:
			logger.V("received push event")

			if evt.UserEmail == m.gitEmail {
				return ""
			}

			return evt.Project.PathWithNamespace
		default:
			logger.V("received ignored event")
			return ""
		}
	}()
	if repo == "" {
		logger.I("no execution triggered")
		w.WriteHeader(http.StatusOK)
		return
	}

	logger = logger.WithFields(log.String("repo", repo))

	logger.I("executing renovate")

	err = m.executor.Execute(types.ExecutionArgs{
		Platform: "gitlab",
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
