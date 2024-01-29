#!/bin/sh
set -e
go test -v -tags=structr_32bit_hash .
go test -v -tags=structr_48bit_hash .
go test -v -tags=structr_64bit_hash .