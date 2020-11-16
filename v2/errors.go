package silverdile

import (
	"fmt"

	"golang.org/x/xerrors"
)

// ErrNotFound is 見つからなかった時に返す
var ErrNotFound = &Error{
	Code:    "NotFound",
	Message: "not found",
	KV:      map[string]interface{}{},
}

// ErrInvalidArgument is 引数に問題がある時に返す
var ErrInvalidArgument = &Error{
	Code:    "InvalidArgument",
	Message: "invalid argument",
	KV:      map[string]interface{}{},
}

// ErrCloudStorage is Cloud Storage への API Request でエラーが発生した時に返す
var ErrCloudStorage = &Error{
	Code:    "CloudStorage",
	Message: "failed cloud storage access",
	KV:      map[string]interface{}{},
}

// ErrNeedConvert is Cloud Storage に キャッシュが見つからず、変換を行う必要がある時に返す
var ErrNeedConvert = &Error{
	Code:    "NeedConvert",
	Message: "need to convert the object",
	KV:      map[string]interface{}{},
}

// ErrInternalError is 何らかのエラーが発生した時に返す
var ErrInternalError = &Error{
	Code:    "InternalError",
	Message: "internal error",
	KV:      map[string]interface{}{},
}

// Error is Error情報を保持する struct
type Error struct {
	Code    string
	Message string
	KV      map[string]interface{}
	err     error
}

// Error is error interface func
func (e *Error) Error() string {
	if e.KV == nil || len(e.KV) < 1 {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: attribute:%+v", e.Code, e.Message, e.KV)
}

// Is is err equal check
func (e *Error) Is(target error) bool {
	var appErr *Error
	if !xerrors.As(target, &appErr) {
		return false
	}
	return e.Code == appErr.Code
}

// Unwrap is return unwrap error
func (e *Error) Unwrap() error {
	return e.err
}

// NewErrNotFound is return ErrNotFound
func NewErrNotFound(key string, err error) error {
	return &Error{
		Code:    ErrNotFound.Code,
		Message: ErrNotFound.Message,
		KV: map[string]interface{}{
			"Target": key,
		},
		err: err,
	}
}

// NewErrInvalidArgument is return InvalidArgument
func NewErrInvalidArgument(message string, kv map[string]interface{}, err error) error {
	return &Error{
		Code:    ErrInvalidArgument.Code,
		Message: message,
		KV:      kv,
		err:     err,
	}
}

// NewErrCloudStorage is return ErrCloudStorage
func NewErrCloudStorage(message string, kv map[string]interface{}, err error) error {
	return &Error{
		Code:    ErrCloudStorage.Code,
		Message: message,
		KV:      kv,
		err:     err,
	}
}

// NewErrNeedConvert is return ErrNeedConvert
func NewErrNeedConvert(message string, kv map[string]interface{}) error {
	return &Error{
		Code:    ErrNeedConvert.Code,
		Message: message,
		KV:      kv,
	}
}

// NewErrInternalError is return ErrInternalError
func NewErrInternalError(message string, kv map[string]interface{}, err error) error {
	return &Error{
		Code:    ErrInternalError.Code,
		Message: message,
		KV:      kv,
		err:     err,
	}
}
