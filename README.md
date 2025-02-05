# Go Logging Library

`glog` is a powerful and flexible logging package for Go, built on top of the standard `log/slog` package. It provides additional features such as customizable output formats, file logging, memory statistics logging, and HTTP request logging middleware.

## Features

- Support for output formats: JSON, TEXT.
- Logging to a file or standard output.
- Context support for passing loggers between functions.
- Flexible configuration of log levels and source addition.
- Middleware for logging HTTP requests.
- Helper for periodic memory statistics logging.

## Installation

To install the package, use the following command:

```bash
go get github.com/kda47/glog
```

## Usage

Creating a Logger

```go
package main

import (
    "github.com/kda47/glog"
)

func main() {
    logger := glog.NewLogger(
    		glog.WithLevel("info"),
        glog.WithAddSource(false),
        glog.WithOutputFormat(glog.OutputFormatTEXT),
    )

    logger.Info("This is an informational message")

    logger.Info("Info", glog.IntAttr("value", 15), glog.StringAttr("status", "finished"))
}
```

HTTP Request Logging Middleware

```go
package main

import (
	"net/http"

	"github.com/kda47/glog"
)

func main() {
	glog.NewLogger(
		glog.WithLevel("info"),
		glog.WithAddSource(false),
		glog.WithOutputFormat(glog.OutputFormatTEXT),
		glog.WithSetDefault(true),
	)
	middleware := glog.NewHttpAccessLogMiddleware("http-access")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!"))
	})

	http.Handle("/", middleware(handler))

	http.ListenAndServe(":8080", nil)
}
```

Debug requests

```go
package main

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/kda47/glog"
)

const (
	loggingDebugHeaderKey   = "X-Log-Level-Debug-Enable"
	loggingDebugSecretToken = "my-super-debug-secret"
)

var defaultLoggerOptions = []glog.LoggerOption{
	glog.WithAddSource(false),
	glog.WithOutputFormat(glog.OutputFormatTEXT),
}

func debugRequestsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		debug := r.Header.Get(loggingDebugHeaderKey)
		ctx := r.Context()
		if debug == loggingDebugSecretToken {
			r.Header.Del(loggingDebugHeaderKey)
			opts := slices.Concat(
				nil,
				[]glog.LoggerOption{glog.WithSetDefault(false), glog.WithLevel("debug")},
				defaultLoggerOptions,
			)
			logger := glog.NewLogger(opts...)
			ctx = glog.ContextWithLogger(ctx, logger)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	logger := glog.L(r.Context())

	var headers []string
	for name, values := range r.Header {
		name = strings.ToLower(name)
		for _, h := range values {
			headers = append(headers, fmt.Sprintf("%v: %v", name, h))
		}
	}
	logger.Debug("Debug message", glog.StringAttr("headers", strings.Join(headers, "; ")))

	w.Write([]byte("Hello, world!"))
}

func main() {
	opts := slices.Concat(
		nil,
		[]glog.LoggerOption{glog.WithSetDefault(true), glog.WithLevel("info")},
		defaultLoggerOptions,
	)
	glog.NewLogger(opts...)

	logRequestsMiddleware := glog.NewHttpAccessLogMiddleware("http-access")

	http.Handle("/", logRequestsMiddleware(debugRequestsMiddleware((http.HandlerFunc(rootHandler)))))

	glog.Default().Info("HTTP Server started", glog.StringAttr("listen", ":"), glog.IntAttr("port", 8080))

	http.ListenAndServe(":8080", nil)
}

```

try to open http://127.0.0.1:88080/ from browser

If you want to debug query call

```bash
curl -H "X-Log-Level-Debug-Enable: my-super-debug-secret" -H "Test: Hello" -X GET http://127.0.0.1:8080/
```


## Testing

Simple run tests
```bash
go test ./...
```

## License
This project is licensed under the MIT License. See the LICENSE file for details.
