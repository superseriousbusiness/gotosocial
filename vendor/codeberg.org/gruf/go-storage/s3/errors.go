package s3

import (
	"strings"

	"codeberg.org/gruf/go-storage"
	"codeberg.org/gruf/go-storage/internal"
	"github.com/minio/minio-go/v7"
)

// transformS3Error transforms an error returned from S3Storage underlying
// minio.Core client, by wrapping where necessary with our own error types.
func transformS3Error(err error) error {
	// Cast this to a minio error response
	ersp, ok := err.(minio.ErrorResponse)
	if ok {
		switch ersp.Code {
		case "NoSuchKey":
			return internal.WrapErr(err, storage.ErrNotFound)
		case "Conflict":
			return internal.WrapErr(err, storage.ErrAlreadyExists)
		default:
			return err
		}
	}

	// Check if error has an invalid object name prefix
	if strings.HasPrefix(err.Error(), "Object name ") {
		return internal.WrapErr(err, storage.ErrInvalidKey)
	}

	return err
}

func isNotFoundError(err error) bool {
	errRsp, ok := err.(minio.ErrorResponse)
	return ok && errRsp.Code == "NoSuchKey"
}

func isConflictError(err error) bool {
	errRsp, ok := err.(minio.ErrorResponse)
	return ok && errRsp.Code == "Conflict"
}

func isObjectNameError(err error) bool {
	return strings.HasPrefix(err.Error(), "Object name ")
}
