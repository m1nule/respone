package respone

import (
	"context"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

const TraceHeader = "X-Trace-Id"

// Body 统一响应体
type Body struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
	Trace string      `json:"trace,omitempty"`
}

// ResponseCtx 带 context 的统一响应入口，自动提取 trace 写入响应头和响应体
func ResponseCtx(ctx context.Context, w http.ResponseWriter, data interface{}, err error) {
	traceID := TraceIDFromCtx(ctx)
	if traceID != "" {
		w.Header().Set(TraceHeader, traceID)
	}

	if err == nil {
		writeJSON(w, http.StatusOK, &Body{
			Code:  CodeSuccess,
			Msg:   "ok",
			Data:  data,
			Trace: traceID,
		})
		return
	}

	if ce, ok := IsCodeError(err); ok {
		writeJSON(w, http.StatusOK, &Body{
			Code:  ce.Code,
			Msg:   ce.Msg,
			Data:  struct{}{},
			Trace: traceID,
		})
		return
	}

	writeJSON(w, http.StatusOK, &Body{
		Code:  CodeServer,
		Msg:   err.Error(),
		Data:  struct{}{},
		Trace: traceID,
	})
}

// Response 统一响应入口（不含 trace）
//   - err == nil: 返回成功，data 为响应数据
//   - err 为 *CodeError: 使用自定义错误码和消息
//   - err 为其他 error: 使用默认服务器错误码
func Response(w http.ResponseWriter, data interface{}, err error) {
	if err == nil {
		writeJSON(w, http.StatusOK, &Body{
			Code: CodeSuccess,
			Msg:  "ok",
			Data: data,
		})
		return
	}

	if ce, ok := IsCodeError(err); ok {
		writeJSON(w, http.StatusOK, &Body{
			Code: ce.Code,
			Msg:  ce.Msg,
			Data: struct{}{},
		})
		return
	}

	writeJSON(w, http.StatusOK, &Body{
		Code: CodeServer,
		Msg:  err.Error(),
		Data: struct{}{},
	})
}

// TraceIDFromCtx 从 context 中提取 OpenTelemetry trace ID
func TraceIDFromCtx(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// Ok 无数据的成功响应
func Ok(w http.ResponseWriter) {
	Response(w, struct{}{}, nil)
}

// OkWithData 带数据的成功响应
func OkWithData(w http.ResponseWriter, data interface{}) {
	Response(w, data, nil)
}

// Fail 失败响应
func Fail(w http.ResponseWriter, err error) {
	Response(w, nil, err)
}

// FailWithCode 指定错误码的失败响应
func FailWithCode(w http.ResponseWriter, code int, msg string) {
	Response(w, nil, NewCodeError(code, msg))
}

func writeJSON(w http.ResponseWriter, statusCode int, body *Body) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(body)
}
