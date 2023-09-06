package micro

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
)

// Route - represent the route for mux
type Route struct {
	Method  string
	Pattern runtime.Pattern
	Handler runtime.HandlerFunc
}

// PathPattern - return a pattern which matches exactly with the path
func PathPattern(path string) runtime.Pattern {
	return runtime.MustPattern(runtime.NewPattern(1, []int{int(utilities.OpLitPush), 0}, []string{path}, ""))
}

// AllPattern - return a pattern which matches any url
func AllPattern() runtime.Pattern {
	return runtime.MustPattern(runtime.NewPattern(1, []int{int(utilities.OpPush), 0}, []string{""}, ""))
}
