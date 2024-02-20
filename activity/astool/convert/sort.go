package convert

import (
	"github.com/superseriousbusiness/activity/astool/gen"
)

// sortableTypeGenerator is a TypeGenerator slice sorted by TypeName.
type sortableTypeGenerator []*gen.TypeGenerator

// Len is the length of this slice.
func (s sortableTypeGenerator) Len() int {
	return len(s)
}

// Less returns true if the TypeName at one index is less than one at another
// index.
func (s sortableTypeGenerator) Less(i, j int) bool {
	return s[i].TypeName() < s[j].TypeName()
}

// Swap elements at indicated indices.
func (s sortableTypeGenerator) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// sortableFuncPropertyGenerator is a FunctionalPropertyGenerator slice sorted
// by PropertyName.
type sortableFuncPropertyGenerator []*gen.FunctionalPropertyGenerator

// Len is the length of this slice.
func (s sortableFuncPropertyGenerator) Len() int {
	return len(s)
}

// Less returns true if the PropertyName at one index is less than one at
// another index.
func (s sortableFuncPropertyGenerator) Less(i, j int) bool {
	return s[i].PropertyName() < s[j].PropertyName()
}

// Swap elements at indicated indices.
func (s sortableFuncPropertyGenerator) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// sortableNonFuncPropertyGenerator is a NonFunctionalPropertyGenerator slice
// sorted by PropertyName.
type sortableNonFuncPropertyGenerator []*gen.NonFunctionalPropertyGenerator

// Len is the length of this slice.
func (s sortableNonFuncPropertyGenerator) Len() int {
	return len(s)
}

// Less returns true if the PropertyName at one index is less than one at
// another index.
func (s sortableNonFuncPropertyGenerator) Less(i, j int) bool {
	return s[i].PropertyName() < s[j].PropertyName()
}

// Swap elements at indicated indices.
func (s sortableNonFuncPropertyGenerator) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// sortablePropertyGenerator is a PropertyGenerator slice sorted by
// PropertyName.
type sortablePropertyGenerator []*gen.PropertyGenerator

// Len is the length of this slice.
func (s sortablePropertyGenerator) Len() int {
	return len(s)
}

// Less returns true if the PropertyName at one index is less than one at
// another index.
func (s sortablePropertyGenerator) Less(i, j int) bool {
	return s[i].PropertyName() < s[j].PropertyName()
}

// Swap elements at indicated indices.
func (s sortablePropertyGenerator) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
