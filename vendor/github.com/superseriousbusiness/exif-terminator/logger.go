/*
   exif-terminator
   Copyright (C) 2022 SuperSeriousBusiness admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package terminator

import "fmt"

var logger ErrorLogger

func init() {
	logger = &defaultErrorLogger{}
}

// ErrorLogger denotes a generic error logging function.
type ErrorLogger interface {
	Error(args ...interface{})
}

type defaultErrorLogger struct{}

func (d *defaultErrorLogger) Error(args ...interface{}) {
	fmt.Println(args...)
}

// SetErrorLogger allows a user of the exif-terminator library
// to set the logger that will be used for error logging.
//
// If it is not set, the default error logger will be used, which
// just prints errors to stdout.
func SetErrorLogger(errorLogger ErrorLogger) {
	logger = errorLogger
}
