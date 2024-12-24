package glog

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AuthInfo struct {
	User string
	Role string
}

func (ai AuthInfo) LogValue() Value {
	return slog.GroupValue(
		StringAttr("user", ai.User),
		StringAttr("role", ai.Role),
	)
}

func NewRecordsLogger(records *[]Record) *Logger { return New(NewRecordsHandler(records)) }

func NewRecordsHandler(records *[]Record) *RecordsHandler { return &RecordsHandler{records: records} }

type RecordsHandler struct{ records *[]Record }

func (h *RecordsHandler) Enabled(_ context.Context, _ Level) bool { return true }

func (h *RecordsHandler) Handle(_ context.Context, r Record) error {
	*h.records = append(*h.records, r)
	return nil
}

func (h *RecordsHandler) WithAttrs(_ []Attr) Handler { return h }

func (h *RecordsHandler) WithGroup(_ string) Handler { return h }

func checkLogRecord(r Record, expectLvl Level, expectMsg string, expectAttrs []Attr) error {
	if r.Level != expectLvl {
		return fmt.Errorf("expected %s level, got %s", expectLvl.String(), r.Level.String())
	}
	if r.Message != expectMsg {
		return fmt.Errorf("expected message '%s', got '%s'", expectMsg, r.Message)
	}
	if r.NumAttrs() < len(expectAttrs) {
		return fmt.Errorf("expected %d attributes, got %d", len(expectAttrs), r.NumAttrs())
	}

	recordAttrs := make(map[string]Value)
	r.Attrs(func(attr slog.Attr) bool {
		recordAttrs[attr.Key] = attr.Value.Resolve()
		return true
	})

	// value := attr.Value.Resolve()
	// switch value.Kind() {
	// case KindGroup:
	// 	for _, groupAttr := range value.Group() {
	// 		fmt.Println(groupAttr.Key, groupAttr.Value)
	// 	}
	// 	recordAttrs[attr.Key] = ""
	// default:
	// 	recordAttrs[attr.Key] = attr.Value.Resolve().String()
	// }

	for _, expectAttr := range expectAttrs {
		expectValue := expectAttr.Value.String()
		recordValue, ok := recordAttrs[expectAttr.Key]
		if !ok {
			return fmt.Errorf("log record doesn't have attr with key '%s'", expectAttr.Key)
		}
		if expectValue != recordValue.String() {
			return fmt.Errorf("got unexpected value '%s' for key '%s', expected '%s'", recordValue.String(), expectAttr.Key, expectValue)
		}
	}
	return nil
}

var testHumanizeUnitsCases = []struct {
	in  int
	out string
}{
	{10, "10 B"},
	{10240, "10.0 KiB"},
	{10485760, "10.0 MiB"},
	{10737418240, "10.0 GiB"},
	{1073741824000, "1000.0 GiB"},
	{10995116277760, "10.0 TiB"},
}

func TestHumanizeUnits(t *testing.T) {
	for _, testCase := range testHumanizeUnitsCases {
		t.Run(testCase.out, func(t *testing.T) {
			pretty := byteCountIEC(testCase.in)
			if pretty != testCase.out {
				t.Errorf("got %q, want %q", pretty, testCase.out)
			}
		})
	}
}

func TestHttpAccessLogMiddleware(t *testing.T) {
	var logRecords []Record

	logger := NewRecordsLogger(&logRecords)
	ctx := ContextWithLogger(context.Background(), logger)
	httpMiddleware := NewHttpAccessLogMiddleware("access")

	// OK response without body
	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	req := httptest.NewRequest("GET", "http://testing", nil).WithContext(ctx)
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	w := httptest.NewRecorder()
	httpMiddleware(http.HandlerFunc(httpHandler)).ServeHTTP(w, req)

	if count := len(logRecords); count != 1 {
		t.Errorf("excepted 1 log record, got %d", count)
	}

	err := checkLogRecord(
		logRecords[0],
		LevelInfo,
		"Request",
		[]Attr{
			StringAttr("ip", host),
			StringAttr("method", "GET"),
			IntAttr("status", 200),
			StringAttr("query", "/"),
			StringAttr("size", "0 B"),
			IntAttr("length", 0),
			StringAttr("agent", req.UserAgent()),
			StringAttr("referer", req.Referer()),
		},
	)
	if err != nil {
		t.Errorf("check log record error: %s", err.Error())
	}

	logRecords = logRecords[:0]

	// 405 response without body
	httpHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	req = httptest.NewRequest("GET", "http://testing", nil).WithContext(ctx)
	host, _, _ = net.SplitHostPort(req.RemoteAddr)
	w = httptest.NewRecorder()
	httpMiddleware(http.HandlerFunc(httpHandler)).ServeHTTP(w, req)

	if count := len(logRecords); count != 1 {
		t.Errorf("excepted 1 log record, got %d", count)
	}

	err = checkLogRecord(
		logRecords[0],
		LevelWarn,
		"Request",
		[]Attr{
			StringAttr("ip", host),
			StringAttr("method", "GET"),
			IntAttr("status", 405),
			StringAttr("query", "/"),
			StringAttr("size", "0 B"),
			IntAttr("length", 0),
			StringAttr("agent", req.UserAgent()),
			StringAttr("referer", req.Referer()),
		},
	)
	if err != nil {
		t.Errorf("check log record error: %s", err.Error())
	}

	logRecords = logRecords[:0]

	// 502 response without body
	httpHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}
	req = httptest.NewRequest("GET", "http://testing", nil).WithContext(ctx)
	host, _, _ = net.SplitHostPort(req.RemoteAddr)
	w = httptest.NewRecorder()
	httpMiddleware(http.HandlerFunc(httpHandler)).ServeHTTP(w, req)

	if count := len(logRecords); count != 1 {
		t.Errorf("excepted 1 log record, got %d", count)
	}

	err = checkLogRecord(
		logRecords[0],
		LevelError,
		"Request",
		[]Attr{
			StringAttr("ip", host),
			StringAttr("method", "GET"),
			IntAttr("status", 502),
			StringAttr("query", "/"),
			StringAttr("size", "0 B"),
			IntAttr("length", 0),
			StringAttr("agent", req.UserAgent()),
			StringAttr("referer", req.Referer()),
		},
	)
	if err != nil {
		t.Errorf("check log record error: %s", err.Error())
	}

	logRecords = logRecords[:0]

	// 201 response with body
	httpHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("hello"))
	}
	req = httptest.NewRequest("POST", "http://testing/cats", nil).WithContext(ctx)
	req.Header.Add("User-Agent", "my-test-agent")
	req.Header.Add("Referer", "http://testing/auth")
	host, _, _ = net.SplitHostPort(req.RemoteAddr)
	w = httptest.NewRecorder()
	httpMiddleware(http.HandlerFunc(httpHandler)).ServeHTTP(w, req)

	if count := len(logRecords); count != 1 {
		t.Errorf("excepted 1 log record, got %d", count)
	}

	err = checkLogRecord(
		logRecords[0],
		LevelInfo,
		"Request",
		[]Attr{
			StringAttr("ip", host),
			StringAttr("method", "POST"),
			IntAttr("status", 201),
			StringAttr("query", "/cats"),
			StringAttr("size", "5 B"),
			IntAttr("length", 5),
			StringAttr("agent", "my-test-agent"),
			StringAttr("referer", "http://testing/auth"),
		},
	)
	if err != nil {
		t.Errorf("check log record error: %s", err.Error())
	}

	logRecords = logRecords[:0]

	// 200 response with empty body
	auth := AuthInfo{User: "admin@test.go", Role: "admin"}
	httpHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(""))
	}
	ctx = ContextWithLoggedHttpAuthInfo(ctx, auth)
	req = httptest.NewRequest("GET", "http://testing/parts", nil).WithContext(ctx)
	req.Header.Add("User-Agent", "my-test-agent")
	req.Header.Add("Referer", "http://testing/auth")
	host, _, _ = net.SplitHostPort(req.RemoteAddr)
	w = httptest.NewRecorder()
	httpMiddleware(http.HandlerFunc(httpHandler)).ServeHTTP(w, req)

	if count := len(logRecords); count != 1 {
		t.Errorf("excepted 1 log record, got %d", count)
	}

	err = checkLogRecord(
		logRecords[0],
		LevelInfo,
		"Request",
		[]Attr{
			StringAttr("name", "access"),
			StringAttr("ip", host),
			StringAttr("method", "GET"),
			IntAttr("status", 200),
			StringAttr("query", "/parts"),
			StringAttr("size", "0 B"),
			IntAttr("length", 0),
			StringAttr("agent", "my-test-agent"),
			StringAttr("referer", "http://testing/auth"),
			Group("auth", StringAttr("user", "admin@test.go"), StringAttr("role", "admin")),
		},
	)
	if err != nil {
		t.Errorf("check log record error: %s", err.Error())
	}
}
