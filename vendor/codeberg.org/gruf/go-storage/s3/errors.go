package s3

import (
	"strings"

	"github.com/minio/minio-go/v7"
)

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
