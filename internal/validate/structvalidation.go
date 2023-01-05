/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package validate

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

var v *validator.Validate

func ulidValidator(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return regexes.ULID.MatchString(field.String())
	default:
		return false
	}
}

func init() {
	v = validator.New()
	if err := v.RegisterValidation("ulid", ulidValidator); err != nil {
		panic(err)
	}
}

// Struct validates the passed struct, returning validator.ValidationErrors if invalid, or nil if OK.
func Struct(s interface{}) error {
	return processValidationError(v.Struct(s))
}

func processValidationError(err error) error {
	if err == nil {
		return nil
	}

	if ive, ok := err.(*validator.InvalidValidationError); ok {
		panic(ive)
	}

	valErr, ok := err.(validator.ValidationErrors)
	if !ok {
		panic("*validator.InvalidValidationError could not be coerced to validator.ValidationErrors")
	}

	return valErr
}
