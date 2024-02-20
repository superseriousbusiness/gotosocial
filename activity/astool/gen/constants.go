package gen

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"strings"
)

// GenerateConstants generates string constants for the type and property
// names. Note that the properties that could be maps have an additional
// constant generatred.
func GenerateConstants(types []*TypeGenerator, props []*PropertyGenerator) (c []jen.Code) {
	for _, t := range types {
		c = append(c,
			jen.Commentf(
				"%s%sName is the string literal of the name for the %s type in the %s vocabulary.", t.VocabName(), t.TypeName(), t.TypeName(), t.VocabName(),
			).Line().Var().Id(
				fmt.Sprintf("%s%sName", t.VocabName(), t.TypeName()),
			).String().Op("=").Lit(t.TypeName()))
	}
	for _, p := range props {
		c = append(c,
			jen.Commentf(
				"%s%sPropertyName is the string literal of the name for the %s property in the %s vocabulary.", p.VocabName(), strings.Title(p.PropertyName()), p.PropertyName(), p.VocabName(),
			).Line().Var().Id(
				fmt.Sprintf("%s%sPropertyName", p.VocabName(), strings.Title(p.PropertyName())),
			).String().Op("=").Lit(p.PropertyName()))
		if p.HasNaturalLanguageMap() {
			c = append(c,
				jen.Commentf(
					"%s%sPropertyMapName is the string literal of the name for the %s property in the %s vocabulary when it is a natural language map.", p.VocabName(), strings.Title(p.PropertyName()), p.PropertyName(), p.VocabName(),
				).Line().Var().Id(
					fmt.Sprintf("%s%sPropertyMapName", p.VocabName(), strings.Title(p.PropertyName())),
				).String().Op("=").Lit(p.PropertyName()+"Map"))
		}
	}
	return
}
