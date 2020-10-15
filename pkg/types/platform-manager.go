package types

import "net/http"

type PlatformManager interface {
	http.Handler
	ListRepos() ([]string, error)
	ExecutionArgs(repos ...string) ExecutionArgs
}
