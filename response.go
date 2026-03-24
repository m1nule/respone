package respone

import (
	"encoding/json"
	"net/http"
)

// Body 统一响应体
type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Response 统一响应入口
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

	// 尝试提取自定义业务错误码
	if ce, ok := IsCodeError(err); ok {
		writeJSON(w, http.StatusOK, &Body{
			Code: ce.Code,
			Msg:  ce.Msg,
			Data: struct{}{},
		})
		return
	}

	// 未知错误，使用默认错误码
	writeJSON(w, http.StatusOK, &Body{
		Code: CodeServer,
		Msg:  err.Error(),
		Data: struct{}{},
	})
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
