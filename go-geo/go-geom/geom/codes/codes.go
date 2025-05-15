package codes

import "errors"

var (
	ErrInconsistentLayout = errors.New("inconsistent layout: input and target layouts don't match")
	ErrUnsupportedLayout  = errors.New("this method does not support the provided layout")
)
