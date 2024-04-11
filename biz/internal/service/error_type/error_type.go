package error_type

import "fmt"

type ErrorType struct {
	Code    int
	Message string
}

func (e *ErrorType) Error() string {
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

var (
	ErrBadRequest   = &ErrorType{Code: 4001, Message: "参数错误"}
	ErrNotFound     = &ErrorType{Code: 4004, Message: "未找到相关数据"}
	ErrNotLogin     = &ErrorType{Code: 4003, Message: "您还未登录"}
	ErrNotPermitted = &ErrorType{Code: 4103, Message: "您无权执行此操作"}
	ErrInternal     = &ErrorType{Code: 5001, Message: "系统错误，请联系管理员"}
)
