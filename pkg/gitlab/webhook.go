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
			repo := evt.Project.PathWithNamespace
			logger = logger.WithFields(log.String("repo", repo))
			logger.V("received issue event")

			expectedTitle := m.getDashboardTitle(repo)

			if expectedTitle == "" {
				// no dashboard issue title provided, we may assume any issue with any title can trigger
				// if they have checkbox (todo list)
			} else if expectedTitle != evt.ObjectAttributes.Title {
				logger.D("issue event is not related to renovate dashboard issue",
					log.String("expected", expectedTitle),
					log.String("actual", evt.ObjectAttributes.Title),
				)
				return ""
			}

			logger.D("checking issue checkbox state")
			if util.CountCheckedItems(evt.ObjectAttributes.Description) > 0 {
				return repo
			}

			return ""
		case *gitlab.MergeEvent:
			repo := evt.Project.PathWithNamespace
			logger = logger.WithFields(log.String("repo", repo))

			logger.V("received merge request event")
			logger.D("checking merge request checkbox state")
			if util.CountCheckedItems(evt.ObjectAttributes.Description) > 0 {
				return repo
			}

			return ""
		case *gitlab.PushEvent:
			repo := evt.Project.PathWithNamespace
			logger = logger.WithFields(log.String("repo", repo))

			logger.V("received push event")

			if evt.UserEmail == m.gitEmail {
				return ""
			}

			return evt.Project.PathWithNamespace
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
