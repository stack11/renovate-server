package types

import "net/http"

type PlatformManager interface {
	http.Handler
}
