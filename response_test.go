package respone

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestResponse_Success(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"name": "test"}

	Response(w, data, nil)

	var body Body
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Code != CodeSuccess {
		t.Errorf("code = %d, want %d", body.Code, CodeSuccess)
	}
	if body.Msg != "ok" {
		t.Errorf("msg = %q, want %q", body.Msg, "ok")
	}
}

func TestResponse_CodeError(t *testing.T) {
	w := httptest.NewRecorder()

	Response(w, nil, NewCodeError(CodeParam, "参数不合法"))

	var body Body
	json.NewDecoder(w.Body).Decode(&body)
	if body.Code != CodeParam {
		t.Errorf("code = %d, want %d", body.Code, CodeParam)
	}
	if body.Msg != "参数不合法" {
		t.Errorf("msg = %q, want %q", body.Msg, "参数不合法")
	}
}

func TestResponse_WrappedCodeError(t *testing.T) {
	w := httptest.NewRecorder()
	original := NewAuthError("token过期")
	wrapped := fmt.Errorf("handler: %w", original)

	Response(w, nil, wrapped)

	var body Body
	json.NewDecoder(w.Body).Decode(&body)
	if body.Code != CodeAuth {
		t.Errorf("code = %d, want %d", body.Code, CodeAuth)
	}
}

func TestResponse_GenericError(t *testing.T) {
	w := httptest.NewRecorder()

	Response(w, nil, errors.New("unexpected"))

	var body Body
	json.NewDecoder(w.Body).Decode(&body)
	if body.Code != CodeServer {
		t.Errorf("code = %d, want %d", body.Code, CodeServer)
	}
}

func TestOk(t *testing.T) {
	w := httptest.NewRecorder()
	Ok(w)

	var body Body
	json.NewDecoder(w.Body).Decode(&body)
	if body.Code != CodeSuccess {
		t.Errorf("code = %d, want %d", body.Code, CodeSuccess)
	}
}

func TestFailWithCode(t *testing.T) {
	w := httptest.NewRecorder()
	FailWithCode(w, CodeForbid, "无权限")

	var body Body
	json.NewDecoder(w.Body).Decode(&body)
	if body.Code != CodeForbid {
		t.Errorf("code = %d, want %d", body.Code, CodeForbid)
	}
}

func TestContentType(t *testing.T) {
	w := httptest.NewRecorder()
	Ok(w)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestHTTPStatus(t *testing.T) {
	w := httptest.NewRecorder()
	Response(w, nil, NewCodeError(CodeParam, "bad param"))

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// --- ResponseCtx tests ---

// newTracedCtx 创建携带 trace ID 的 context（用于测试）
func newTracedCtx(traceHex string) context.Context {
	tid, _ := trace.TraceIDFromHex(traceHex)
	sid, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	})
	return trace.ContextWithSpanContext(context.Background(), sc)
}

func TestResponseCtx_TraceInHeader(t *testing.T) {
	w := httptest.NewRecorder()
	traceID := "4f96e569df4c0a64b4c55190e44bd94e"
	ctx := newTracedCtx(traceID)

	ResponseCtx(ctx, w, map[string]string{"ok": "1"}, nil)

	got := w.Header().Get(TraceHeader)
	if got != traceID {
		t.Errorf("header %s = %q, want %q", TraceHeader, got, traceID)
	}
}

func TestResponseCtx_TraceInBody(t *testing.T) {
	w := httptest.NewRecorder()
	traceID := "4f96e569df4c0a64b4c55190e44bd94e"
	ctx := newTracedCtx(traceID)

	ResponseCtx(ctx, w, nil, nil)

	var body Body
	json.NewDecoder(w.Body).Decode(&body)
	if body.Trace != traceID {
		t.Errorf("body.Trace = %q, want %q", body.Trace, traceID)
	}
	if body.Code != CodeSuccess {
		t.Errorf("code = %d, want %d", body.Code, CodeSuccess)
	}
}

func TestResponseCtx_ErrorWithTrace(t *testing.T) {
	w := httptest.NewRecorder()
	traceID := "abcdef1234567890abcdef1234567890"
	ctx := newTracedCtx(traceID)

	ResponseCtx(ctx, w, nil, NewCodeError(CodeAuth, "token无效"))

	got := w.Header().Get(TraceHeader)
	if got != traceID {
		t.Errorf("header %s = %q, want %q", TraceHeader, got, traceID)
	}
	var body Body
	json.NewDecoder(w.Body).Decode(&body)
	if body.Trace != traceID {
		t.Errorf("body.Trace = %q, want %q", body.Trace, traceID)
	}
	if body.Code != CodeAuth {
		t.Errorf("code = %d, want %d", body.Code, CodeAuth)
	}
}

func TestResponseCtx_NoTrace(t *testing.T) {
	w := httptest.NewRecorder()

	ResponseCtx(context.Background(), w, nil, nil)

	got := w.Header().Get(TraceHeader)
	if got != "" {
		t.Errorf("header %s should be empty, got %q", TraceHeader, got)
	}
	var body Body
	json.NewDecoder(w.Body).Decode(&body)
	if body.Trace != "" {
		t.Errorf("body.Trace should be empty, got %q", body.Trace)
	}
}
