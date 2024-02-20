package gen

import (
	"fmt"

	"github.com/superseriousbusiness/activity/astool/codegen"
)

func GenRootPackageComment(pkgName string) string {
	return codegen.FormatPackageDocumentation(fmt.Sprintf("Package %s "+
		"contains constructors and functions necessary for "+
		"applications to serialize, deserialize, and use "+
		"ActivityStreams types in Go. This package is code-generated "+
		"and subject to the same license as the go-fed tool used to "+
		"generate it.\n\n"+
		"This package is useful to three classes of developers: "+
		"end-user-application developers, specification writers "+
		"creating an ActivityStream Extension, and ActivityPub "+
		"implementors wanting to create an alternate ActivityStreams "+
		"implementation that still satisfies the interfaces generated "+
		"by the go-fed tool.\n\n"+
		"Application developers should limit their use to the "+
		"Resolver type, the constructors beginning with \"New\", the "+
		"\"Extends\" functions, the \"DisjointWith\" functions, the "+
		"\"ExtendedBy\" functions, and any interfaces returned in "+
		"those functions in this package. This lets applications use "+
		"Resolvers to Deserialize or Dispatch specific types. The "+
		"types themselves can Serialize as needed. The \"Extends\", "+
		"\"DisjointWith\", and \"ExtendedBy\" functions help navigate "+
		"the ActivityStreams hierarchy since it is not equivalent to "+
		"object-oriented inheritance.\n\n"+
		"When creating an ActivityStreams extension, developers will "+
		"want to ensure that the generated code builds correctly and "+
		"check that the properties, types, extensions, and "+
		"disjointedness is set up correctly. Writing unit tests with "+
		"concrete types is then the next step. If the tool has an "+
		"error generating this code, a fix is needed in the tool as "+
		"it is likely there is a new RDF type being used in the "+
		"extension that the tool does not know how to resolve. Thus, "+
		"most development will focus on the go-fed tool itself."+
		"\n\n"+
		"Finally, ActivityStreams implementors that want drop-in "+
		"replacement while still using the generated interfaces are "+
		"highly encouraged to examine the Manager type in this "+
		"package (in addition to the constructors) as these are the "+
		"locations where concrete types are instantiated. When "+
		"supplying a different type in these two locations, the "+
		"other generated code will propagate it throughout the "+
		"rest of an application. The Manager is instantiated as a "+
		"singleton at init time in this library. It is then injected "+
		"into each implementation library so they can deserialize "+
		"their needed types without relying on the underlying "+
		"concrete type.\n\n"+
		"Subdirectories of this package include implementation "+
		"files and functions that are not intended to be directly "+
		"linked to applications, but are used by this particular "+
		"package. It is strongly recommended to only use the "+
		"property interfaces and type interfaces in subdirectories "+
		"and limiting concrete types to those in this package. The "+
		"go-fed tool is likely to contain a pruning feature in the "+
		"future which will analyze an application and eliminate "+
		"code that would be dead if it were to be generated which "+
		"reduces the compilation time, compilation resources, and "+
		"binary size of an application. Such a feature will not be "+
		"compatible with applications that use the concrete "+
		"implementation types.",
		pkgName))
}

func VocabPackageComment(pkgName, vocabName string) string {
	return codegen.FormatPackageDocumentation(fmt.Sprintf("Package %s "+
		"contains the interfaces for the %s vocabulary. All "+
		"applications are strongly encouraged to use these interface "+
		"types instead of the concrete definitions contained in the "+
		"implementation subpackage. These interfaces allow "+
		"applications to consume only the types and properties "+
		"needed and be independent of the go-fed implementation if "+
		"another alternative implementation is created. This package "+
		"is code-generated and subject to the same license as the "+
		"go-fed tool used to generate it.\n\n"+
		"Type interfaces contain \"Get\" and \"Set\" methods for "+
		"its properties. Types also have a \"Serialize\" method to "+
		"convert the type into an interface map for use with the json "+
		"package. There is a convenience \"IsExtending\" method on "+
		"each types which helps with the ActivityStreams hierarchy, "+
		"which is not the same as object oriented inheritance. While "+
		"types also have a \"LessThan\" method, it is an arbitrary "+
		"sort. Do not use it if needing to sort on specific "+
		"properties, such as publish time. It is best used for "+
		"normalizing the type. Lastly, do not use the "+
		"\"GetUnknownProperties\" method in an application. Instead, "+
		"use the go-fed tool to code generate the property needed. "+
		"\n\n"+
		"Properties come in two flavors: functional and "+
		"non-functional. Functional means that a property can have at "+
		"most one value, while non-functional means a property could "+
		"have zero, one, or more values. Any property value may also "+
		"be an IRI, in which case the application will need to make a "+
		"HTTP request to fetch the property value.\n\n"+
		"Functional properties have \"Get\", \"Is\", and \"Set\" "+
		"methods for determining what kind of value the property is, "+
		"fetching that value, or setting that value. There is also "+
		"a \"Serialize\" method which converts the property into an "+
		"interface type, but applications should not typically use "+
		"a property's \"Serialize\" and instead should use a type's "+
		"\"Serialize\" instead. Like types, properties have an "+
		"arbitrary \"LessThan\" comparison function that should not "+
		"be used if needing to sort on specific values. Finally, "+
		"applications should not use the \"KindIndex\" method as it "+
		"is a comparison mechanism only for those looking to write an "+
		"alternate implementation.\n\n"+
		"Non-functional properties can have more than one value, so "+
		"it has  \"Len\" for getting its length, \"At\" for getting "+
		"an iterator pointing to an element, \"Append\" and "+
		"\"Prepend\" for adding values, \"Remove\" for removing a "+
		"value, \"Set\" for overwriting a value, and \"Swap\" for "+
		"swapping two values' indices. Note that a non-functional "+
		"property satisfies the sort interface, but it results in an "+
		"arbitrary but stable ordering best used as a normalized "+
		"form. A non-functional property's iterator looks like a "+
		"functional property with \"Next\" and \"Previous\" methods. "+
		"Applications should not use the \"KindIndex\" methods as it "+
		"is a comparison mechanism only for those looking to write an "+
		"alternate implementation of this library.\n\n"+
		"Types and properties have a \"JSONLDContext\" method that "+
		"returns a mapping of vocabulary URIs to aliases that are "+
		"required in the JSON-LD @context when serializing this "+
		"value. The aliases used by this library when serializing "+
		"objects is done at code-generation time, unless a different "+
		"alias was used to deserialize the type or property.\n\n"+
		"Types, functional properties, and non-functional properties "+
		"are not designed for concurrent usage by two or more "+
		"goroutines. Also, certain methods on a non-functional "+
		"property will invalidate iterators and possibly cause "+
		"unexpected behaviors. To avoid this, re-obtain an iterator "+
		"after modifying a non-functional property.",
		pkgName, vocabName))
}

func PrivateFlatPackageComment(pkgName, vocabName string) string {
	return codegen.FormatPackageDocumentation(fmt.Sprintf("Package %s "+
		"contains the implementations for the %s vocabulary. All "+
		"applications are strongly encouraged to use the interface "+
		"types instead of these concrete definitions. The interfaces "+
		"allow applications to consume only the types and properties "+
		"needed and be independent of the go-fed implementation if "+
		"another alternative implementation is created. This package "+
		"is code-generated and subject to the same license as the "+
		"go-fed tool used to generate it.\n\n"+
		"This package is independent of other vocabulary "+
		"implementations by having a Manager injected into it to act "+
		"as a factory for the concrete implementations of other "+
		"types. The implementations have been generated together into "+
		"a single implementation library.\n\n"+
		"Strongly consider using the interfaces instead of this "+
		"package.",
		pkgName, vocabName))
}

func PrivateIndividualTypePackageComment(pkgName, typeName string) string {
	return codegen.FormatPackageDocumentation(fmt.Sprintf("Package %s "+
		"contains the implementation for the %s type. All "+
		"applications are strongly encouraged to use the interface "+
		"instead of this concrete definition. The interfaces "+
		"allow applications to consume only the types and properties "+
		"needed and be independent of the go-fed implementation if "+
		"another alternative implementation is created. This package "+
		"is code-generated and subject to the same license as the "+
		"go-fed tool used to generate it.\n\n"+
		"This package is independent of other types' and properties' "+
		"implementations by having a Manager injected into it to act "+
		"as a factory for the concrete implementations. The "+
		"implementations have been generated into their own separate "+
		"subpackages for each vocabulary.\n\n"+
		"Strongly consider using the interfaces instead of this "+
		"package.",
		pkgName, typeName))
}

func PrivateIndividualPropertyPackageComment(pkgName, propertyName string) string {
	return codegen.FormatPackageDocumentation(fmt.Sprintf("Package %s "+
		"contains the implementation for the %s property. All "+
		"applications are strongly encouraged to use the interface "+
		"instead of this concrete definition. The interfaces "+
		"allow applications to consume only the types and properties "+
		"needed and be independent of the go-fed implementation if "+
		"another alternative implementation is created. This package "+
		"is code-generated and subject to the same license as the "+
		"go-fed tool used to generate it.\n\n"+
		"This package is independent of other types' and properties' "+
		"implementations by having a Manager injected into it to act "+
		"as a factory for the concrete implementations. The "+
		"implementations have been generated into their own separate "+
		"subpackages for each vocabulary.\n\n"+
		"Strongly consider using the interfaces instead of this "+
		"package.",
		pkgName, propertyName))
}
