package silverdile

import (
	"github.com/morikuni/failure"
)

var InvalidArgument failure.StringCode = "InvalidArgument"
var NotFound failure.StringCode = "NotFound"
var InternalError failure.StringCode = "InternalError"

func IsErrInvalidArgument(err error) bool {
	return isError(err, InvalidArgument)
}

func IsErrNotFound(err error) bool {
	return isError(err, NotFound)
}

func IsErrInternalError(err error) bool {
	return isError(err, InternalError)
}

func isError(err error, code failure.StringCode) bool {
	v, ok := failure.CodeOf(err)
	if !ok {
		return false
	}
	return v == code
}
