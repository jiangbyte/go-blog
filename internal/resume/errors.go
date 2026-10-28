package resume

import "errors"

var (
	errNotFound   = errors.New("not found")
	errBadRequest = errors.New("bad request")
)

// ErrNotFound 资源不存在。
func ErrNotFound(msg string) error {
	return &bizError{kind: errNotFound, msg: msg}
}

// ErrBadRequest 参数/业务校验失败。
func ErrBadRequest(msg string) error {
	return &bizError{kind: errBadRequest, msg: msg}
}

type bizError struct {
	kind error
	msg  string
}

func (e *bizError) Error() string { return e.msg }
func (e *bizError) Unwrap() error { return e.kind }

func IsNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}

func IsBadRequest(err error) bool {
	return errors.Is(err, errBadRequest)
}
