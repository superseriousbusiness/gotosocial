package internal

import (
	"reflect"
)

func MakeSliceNextElemFunc(v reflect.Value) func() reflect.Value {
	if v.Kind() == reflect.Array {
		var pos int
		return func() reflect.Value {
			v := v.Index(pos)
			pos++
			return v
		}
	}

	elemType := v.Type().Elem()

	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
		return func() reflect.Value {
			if v.Len() < v.Cap() {
				v.Set(v.Slice(0, v.Len()+1))
				elem := v.Index(v.Len() - 1)
				if elem.IsNil() {
					elem.Set(reflect.New(elemType))
				}
				return elem
			}

			elem := reflect.New(elemType)
			v.Set(reflect.Append(v, elem))
			return elem
		}
	}

	zero := reflect.Zero(elemType)
	return func() reflect.Value {
		if v.Len() < v.Cap() {
			v.Set(v.Slice(0, v.Len()+1))
			return v.Index(v.Len() - 1)
		}

		v.Set(reflect.Append(v, zero))
		return v.Index(v.Len() - 1)
	}
}

func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

func FieldByIndexAlloc(v reflect.Value, index []int) reflect.Value {
	if len(index) == 1 {
		return v.Field(index[0])
	}

	for i, idx := range index {
		if i > 0 {
			v = indirectNil(v)
		}
		v = v.Field(idx)
	}
	return v
}

func indirectNil(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

// MakeQueryBytes returns zero-length byte slice with capacity of 4096.
func MakeQueryBytes() []byte {
	// TODO: make this configurable?
	return make([]byte, 0, 4096)
}
