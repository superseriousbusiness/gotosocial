package gen

import ()

const (
	JSONLDVocabName     = "JSONLD"
	JSONLDIdName        = "id"
	JSONLDTypeName      = "type"
	jsonLDIdCamelName   = "Id"
	jsonLDTypeCamelName = "Type"
	jsonLDIdComment     = `Provides the globally unique identifier for JSON-LD entities.`
	jsonLDTypeComment   = `Identifies the schema type(s) of the JSON-LD entity.`
)

// NewIdPropety returns the functional property for the JSON-LD "@id" property.
func NewIdProperty(pm *PackageManager, xsdAnyUri Kind) (*FunctionalPropertyGenerator, error) {
	return NewFunctionalPropertyGenerator(
		JSONLDVocabName,
		nil,
		"",
		pm,
		Identifier{
			LowerName: JSONLDIdName,
			CamelName: jsonLDIdCamelName,
		},
		jsonLDIdComment,
		[]Kind{xsdAnyUri},
		false)
}

// NewTypeProperty returns the non-functional property for the JSON-LD "@type"
// property.
func NewTypeProperty(pm *PackageManager, xsdAnyUri, xsdString Kind) (*NonFunctionalPropertyGenerator, error) {
	return NewNonFunctionalPropertyGenerator(
		JSONLDVocabName,
		nil,
		"",
		pm,
		Identifier{
			LowerName: JSONLDTypeName,
			CamelName: jsonLDTypeCamelName,
		},
		jsonLDTypeComment,
		[]Kind{xsdAnyUri, xsdString},
		false)
}
