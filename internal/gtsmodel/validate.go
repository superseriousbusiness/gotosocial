/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package gtsmodel

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

var v *validator.Validate

const (
	PointerValidationPanic = "validate function was passed pointer"
	InvalidValidationPanic = "validate function was passed invalid item"
)

var ulidValidator = func(fl validator.FieldLevel) bool {
	value, kind, _ := fl.ExtractType(fl.Field())

	if kind != reflect.String {
		return false
	}

	// we want either an empty string, or a proper ULID, nothing else
	// if the string is empty, the `required` tag will take care of it so we don't need to worry about it here
	s := value.String()
	if len(s) == 0 {
		return true
	}
	return util.ValidateULID(s)
}

func init() {
	v = validator.New()
	v.RegisterValidation("ulid", ulidValidator)
}

func ValidateStruct(s interface{}) error {
	switch reflect.ValueOf(s).Kind() {
	case reflect.Invalid:
		panic(InvalidValidationPanic)
	case reflect.Ptr:
		panic(PointerValidationPanic)
	}

	err := v.Struct(s)
	return processValidationError(err)
}

func processValidationError(err error) error {
	if err == nil {
		return nil
	}

	if ive, ok := err.(*validator.InvalidValidationError); ok {
		panic(ive)
	}

	return err.(validator.ValidationErrors)
}
