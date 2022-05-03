package cache

import (
	"reflect"
)

type Comparable interface {
	Equal(any) bool
}

// Compare returns whether 2 values are equal using the Comparable
// interface, or failing that falls back to use reflect.DeepEqual().
func Compare(i1, i2 any) bool {
	c1, ok1 := i1.(Comparable)
	if ok1 {
		return c1.Equal(i2)
	}
	c2, ok2 := i2.(Comparable)
	if ok2 {
		return c2.Equal(i1)
	}
	return reflect.DeepEqual(i1, i2)
}
