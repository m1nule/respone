package respone

import "errors"

// 预定义业务错误码
const (
	CodeSuccess      = 0     // 成功
	CodeServer       = 10001 // 服务器内部错误
	CodeParam        = 10002 // 参数错误
	CodeAuth         = 10003 // 认证失败
	CodeNotFound     = 10004 // 资源不存在
	CodeForbid       = 10005 // 无权限
	CodeTokenExpired = 10006 // token过期
)

// CodeError 自定义业务错误，携带错误码
type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Error 实现 error 接口
func (e *CodeError) Error() string {
	return e.Msg
}

// NewCodeError 创建指定错误码的业务错误
func NewCodeError(code int, msg string) *CodeError {
	return &CodeError{Code: code, Msg: msg}
}

// NewDefaultError 创建默认错误码（10001）的业务错误
func NewDefaultError(msg string) *CodeError {
	return &CodeError{Code: CodeServer, Msg: msg}
}

// NewParamError 创建参数错误
func NewParamError(msg string) *CodeError {
	return &CodeError{Code: CodeParam, Msg: msg}
}

// NewAuthError 创建认证错误
func NewAuthError(msg string) *CodeError {
	return &CodeError{Code: CodeAuth, Msg: msg}
}

// IsCodeError 判断 error 是否为 CodeError 并提取
func IsCodeError(err error) (*CodeError, bool) {
	var ce *CodeError
	if errors.As(err, &ce) {
		return ce, true
	}
	return nil, false
}
