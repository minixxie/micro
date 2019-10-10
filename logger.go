package micro

import jaeger "github.com/uber/jaeger-client-go"

var logger jaeger.Logger

func init() {
	logger = jaeger.NullLogger
}

// SetLogger - set the logger
func SetLogger(l jaeger.Logger) {
	logger = l
}

// Logger - get the logger
func Logger() jaeger.Logger {
	return logger
}
