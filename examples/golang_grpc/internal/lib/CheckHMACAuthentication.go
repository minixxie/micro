package lib

import (
	"net/http"
)

func CheckHMACAuthentication(r *http.Request, customizedCheckFunc func(authHeader string) bool) bool {
	authHeader := r.Header.Get("Authorization")
	return customizedCheckFunc(authHeader)
}
