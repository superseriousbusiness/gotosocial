package exif

import (
	"errors"
)

var (
	ErrTagNotFound    = errors.New("tag not found")
	ErrTagNotStandard = errors.New("tag not a standard tag")
)
