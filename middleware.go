package glog

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	AccessLogNameKey = "name"
	noWritten        = -1
	defaultStatus    = http.StatusOK
)

type loggedHttpAuthInfoContextKey struct{}

func ContextWithLoggedHttpAuthInfo(ctx context.Context, authInfo LogValuer) context.Context {
	return context.WithValue(ctx, loggedHttpAuthInfoContextKey{}, authInfo)
}

func loggedAuthInfoFromContext(ctx context.Context) (LogValuer, bool) {
	auth, ok := ctx.Value(loggedHttpAuthInfoContextKey{}).(LogValuer)
	return auth, ok
}

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	Status() int
	Size() int
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) Status() int { return rw.status }

func (rw *responseWriter) Size() int { return rw.size }

func (rw *responseWriter) WriteHeader(status int) {
	if rw.status == 0 {
		rw.status = status
		rw.ResponseWriter.WriteHeader(status)
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.Status() == 0 {
		rw.WriteHeader(defaultStatus)
	}

	var err error
	rw.size, err = rw.ResponseWriter.Write(data)

	return rw.size, err
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the hijacker interface is not supported")
	}
	return hj.Hijack()
}

func (rw *responseWriter) Flush() {
	if fl, ok := rw.ResponseWriter.(http.Flusher); ok {
		if rw.Status() == 0 {
			rw.WriteHeader(defaultStatus)
		}
		fl.Flush()
	}
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func byteCountIEC(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func getUserIP(r *http.Request) net.IP {
	var userIP string
	if len(r.Header.Get("CF-Connecting-IP")) > 1 {
		userIP = r.Header.Get("CF-Connecting-IP")
		return net.ParseIP(userIP)
	} else if len(r.Header.Get("X-Forwarded-For")) > 1 {
		userIP = r.Header.Get("X-Forwarded-For")
		return net.ParseIP(userIP)
	} else if len(r.Header.Get("X-Real-IP")) > 1 {
		userIP = r.Header.Get("X-Real-IP")
		return net.ParseIP(userIP)
	} else {
		userIP = r.RemoteAddr
		if strings.Contains(userIP, ":") {
			return net.ParseIP(strings.Split(userIP, ":")[0])
		} else {
			return net.ParseIP(userIP)
		}
	}
}

func NewHttpAccessLogMiddleware(name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			status := rw.Status()

			var level Level
			switch {
			case status >= http.StatusInternalServerError:
				level = LevelError
			case status >= http.StatusBadRequest:
				level = LevelWarn
			default:
				level = LevelInfo
			}

			ctx := r.Context()
			logger := L(ctx)

			attrs := []Attr{
				StringAttr(AccessLogNameKey, name),
				StringAttr("method", r.Method),
				StringAttr("ip", getUserIP(r).String()),
				IntAttr("status", status),
				StringAttr("query", r.URL.RequestURI()),
				StringAttr("size", byteCountIEC(rw.Size())),
				IntAttr("length", rw.Size()),
				Float64Attr("duration", time.Since(start).Seconds()),
				StringAttr("agent", r.UserAgent()),
				StringAttr("referer", r.Referer()),
			}

			if authInfo, ok := loggedAuthInfoFromContext(ctx); ok {
				attrs = append(attrs, Any("auth", authInfo))
			}

			logger.LogAttrs(ctx, level, "Request", attrs...)
		})
	}
}
