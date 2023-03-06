package form

import (
	"reflect"
	"strconv"
)

// ExtractType gets the actual underlying type of field value.
// it is exposed for use within you Custom Functions
func ExtractType(current reflect.Value) (reflect.Value, reflect.Kind) {

	switch current.Kind() {
	case reflect.Ptr:

		if current.IsNil() {
			return current, reflect.Ptr
		}

		return ExtractType(current.Elem())

	case reflect.Interface:

		if current.IsNil() {
			return current, reflect.Interface
		}

		return ExtractType(current.Elem())

	default:
		return current, current.Kind()
	}
}

func parseBool(str string) (bool, error) {

	switch str {
	case "1", "t", "T", "true", "TRUE", "True", "on", "yes", "ok":
		return true, nil
	case "", "0", "f", "F", "false", "FALSE", "False", "off", "no":
		return false, nil
	}

	// strconv.NumError mimicing exactly the strconv.ParseBool(..) error and type
	// to ensure compatibility with std library and beyond.
	return false, &strconv.NumError{Func: "ParseBool", Num: str, Err: strconv.ErrSyntax}
}

// hasValue determines if a reflect.Value is it's default value
func hasValue(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		if !field.IsValid() {
			return false
		}
		if !field.Type().Comparable() {
			return true
		}
		return field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}
