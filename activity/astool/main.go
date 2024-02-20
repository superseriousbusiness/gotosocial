package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/superseriousbusiness/activity/astool/convert"
	"github.com/superseriousbusiness/activity/astool/gen"
	"github.com/superseriousbusiness/activity/astool/rdf"
	"github.com/superseriousbusiness/activity/astool/rdf/owl"
	"github.com/superseriousbusiness/activity/astool/rdf/rdfs"
	"github.com/superseriousbusiness/activity/astool/rdf/rfc"
	"github.com/superseriousbusiness/activity/astool/rdf/schema"
	"github.com/superseriousbusiness/activity/astool/rdf/xsd"
)

const (
	pathFlag = "path"
	specFlag = "spec"
	helpText = `
Usage: astool [-spec=<file>] [-path=<gopath prefix>] <directory>

The ActivityStreams tool (astool) is used to generate ActivityStreams types,
properties, and values from an OWL2 RDF specification. The tool generates the
code necessary to create interfaces and functions that solve the problems of
serialization & deserialization of functional and nonfunctional properties,
serialization & deserialization of types, navigating the extends/disjoint
hierarchy, and resolving an arbitrary ActivityStreams into a concrete Go type.

The tool generates files in the current working directory, and creates
subpackages as needed. To generate the code for a specification, pass the OWL
ontology defined as JSON-LD to the tool:

    astool -spec specification.jsonld ./gen/to/subdir

The @context provided in the ActivityStreams specification may be insufficient
for this tool to use to generate code. However, if this tool is able to use the
JSON-LD specification to generate the code, then it should also be compatible
with the @context.

This tool will automatically detect the correct Go prefix path to use if used
in a subdirectory under GOPATH. If used outside of GOPATH, the prefix to the
current working directory must be provided:

    astool -spec specification.jsonld -path path/to/my/module/cwd .

If a specification builds off of a previous specification, they must be provided
in the order of root to dependency, with the ActivityStreams Core & Extended
Types specification as the root:

    astool -spec activitystreams.jsonld -spec derived_extension.jsonld .

The following directories are generated in the current working directory (cwd)
given a particular specification for a <vocabulary>:

    cwd/
        gen_doc.go
            - Package level documentation.
	gen_init.go
	    - Init function definitions.
	gen_manager.go
	    - Definition of Manager, which is responsible for dependency
	      injection of concrete values at runtime for deserialization.
	gen_pkg_<vocabulary>_disjoint.go
	    - Functions determining the "disjointedness" of ActivityStreams
	      types in the specified vocabulary.
	gen_pkg_<vocabulary>_extendedby.go
	    - Functions determining the parent-to-child "extends" of
	      ActivityStreams types in the specified vocabulary.
	gen_pkg_<vocabulary>_extends.go
	    - Functions determining the child-to-parent "extends" of
	      ActivityStreams types in the specified vocabulary.
	gen_pkg_<vocabulary>_property_constructors.go
	    - Constructors of properties in the specified vocabulary.
	gen_pkg_<vocabulary>_type_constructors.go
	    - Constructors of types in the specified vocabulary.

	resolver/
	    gen_type_resolver.go
	        - Resolves arbitrary ActivityStream objects by type.
	    gen_interface_resolver.go
	        - Resolves arbitrary ActivityStream objects by their assertable
		  interfaces.
	    gen_type_predicated_resolver.go
	        - Conditionally resolves based on the ActivityStream object's
		  type.
	    gen_interface_predicated_resolver.go
	        - Conditionally resolves based on the ACtivityStream's
		  assertable interfaces.
	    gen_resolver_utils.go
	        - Functions aiding in handling resolver errors.

	vocab/
	    gen_doc.go
	        - Package level documentation.
	    gen_pkg.go
	        - Generic interface definition.
	    gen_property_<property>_interface.go
	        - Interface definition of a property.
		- NOTE: Application developers should prefer using these
		  interfaces over the concrete types defined in "impl".
	    gen_type_<type>_interface.go
	        - Interface definition of a type.
		- NOTE: Application developers should prefer using these
		  interfaces over the concrete types defined in "impl".

	values/
	    <value>/
	        - Contains RDF values and their serialization, deserialization,
	          and comparison methods.

	impl/
	    <vocabulary>/
	        - Implementation of the vocabulary.
		- NOTE: Application developers should strongly prefer using the
		  interfaces in "vocab" over these.

This tool is geared for three kinds of developers:

1) Application developers can use the tool to generate the native Go types
   needed to build an application.
2) Developers wishing to extend ActivityStreams may use the tool to evaluate
   their OWL definition of their new ActivityStreams types and properties to
   rapidly prototype in Go code.
3) Finally, developers wishing to provide an alternate implementation to go-fed
   can target the same interfaces generated by this tool, and create a fork that
   allows the generated Manager and constructors to inject their concrete type
   into any existing application using go-fed.

The tool relies on built-in knowledge of several ontologies: RDF, RDFS, OWL,
Schema.org, XML, and a few RFCs. However, this tool doesn't have complete
knowledge of all of these ontologies. It may error out because a provided
specification uses a definition that the tool doesn't currently know. In such a
case, please file an issue at https://github.com/go-fed/activity in order to
include the missing definition.

Experimental support for generating the code as a module is provided by settting
the 'path' flag, which will prefix all generated code with the 'path':

    astool -spec specification.jsonld -path mymodule ./subdir

`
)

// Global registry of "known" RDF ontologies. This manages the built-in
// knowledge of how to parse specific linked data documents. It may be cloned
// in the course of processing a JSON-LD document, due to "@context" dictating
// certain ontologies being aliased in some specifications and not others.
var registry *rdf.RDFRegistry

// mustAddOntology ensures that the registry global variable is not nil, and
// then adds the specific ontology or panics if it cannot.
func mustAddOntology(o rdf.Ontology) {
	if registry == nil {
		registry = rdf.NewRDFRegistry()
	}
	if err := registry.AddOntology(o); err != nil {
		panic(err)
	}
}

// At init time, get our built-in knowledge of OWL and other RDF ontologies
// into the registry, before main executes.
func init() {
	flag.Usage = func() {
		_, _ = io.WriteString(flag.CommandLine.Output(), helpText)
		flag.PrintDefaults()
	}
	mustAddOntology(&xsd.XMLOntology{Package: "xml"})
	mustAddOntology(&owl.OWLOntology{})
	mustAddOntology(&rdf.RDFOntology{Package: "rdf"})
	mustAddOntology(&rdfs.RDFSchemaOntology{})
	mustAddOntology(&schema.SchemaOntology{})
	mustAddOntology(&rfc.RFCOntology{Package: "rfc"})
}

// list is a flag-friendly comma-separated list of strings. Also allows multiple
// definitions of the flag to not overwrite each other and instead result in a
// list of strings.
//
// The values of the flag cannot contain commas within them because the value
// will be split into two.
type list []string

// String turns this list into a single comma-separated string.
func (l *list) String() string {
	return strings.Join(*l, ",")
}

// Set adds a string value to the list, after splitting on the comma separator.
func (l *list) Set(v string) error {
	vals := strings.Split(v, ",")
	*l = append(*l, vals...)
	return nil
}

// settableString is a flag-friendly string that distinguishes an empty string
// due to not being set and explicitly being set as empty at the command line.
type settableString struct {
	set bool
	str string
}

// String simply returns the string value of this settableString.
func (s *settableString) String() string {
	return s.str
}

// Set will mark this settableString's set as true and store the value.
func (s *settableString) Set(v string) error {
	s.set = true
	s.str = v
	return nil
}

// IsSet returns true if this value was explicitly set as a flag value.
func (s settableString) IsSet() bool {
	return s.set
}

// CommandLineFlags manages the flags defined by this tool.
type CommandLineFlags struct {
	// Flags
	specs list
	path  settableString
	// Additional data
	pathAutoDetected bool
	// Destination on the file system for the code generation
	destination string
}

// NewCommandLineFlags defines the flags expected to be used by this tool. Calls
// flag.Parse on behalf of the main program, and validates the flags. Returns an
// error if validation fails.
func NewCommandLineFlags() (*CommandLineFlags, error) {
	c := &CommandLineFlags{}
	flag.Var(
		&c.path,
		pathFlag,
		"Package path to use for all generated package paths. If using GOPATH, this is automatically detected as $GOPATH/<path>/ when generating in a subdirectory. Cannot be explicitly set to be empty.")
	flag.Var(&(c.specs), specFlag, "Input JSON-LD specification used to generate Go code.")
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		return nil, fmt.Errorf("astool requires a destination directory")
	}
	c.destination = args[0]
	return c, c.Validate()
}

// detectPath attempts to detect the path to use when generating the code. The
// path is only detected if the tool is running in a subdirectory of GOPATH,
// and will be set to $GOPATH/<path>/. After this method runs without errors,
// c.path.IsSet will always return true.
//
// When auto-detecting, if GOPATH is not set then will return an error.
//
// If the path has already been set at the command line, does nothing.
func (c *CommandLineFlags) detectPath() error {
	if c.path.IsSet() {
		return nil
	}
	gopath, isSet := os.LookupEnv("GOPATH")
	if !isSet {
		return fmt.Errorf("cannot detect %q because GOPATH environmental variable is not set and %q flag was not explicitly set", pathFlag, pathFlag)
	}
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(pwd, gopath) {
		return fmt.Errorf("cannot detect %q because current working directory is not under GOPATH and %q flag was not explicitly set", pathFlag, pathFlag)
	}
	c.pathAutoDetected = true
	gopath = strings.Join([]string{gopath, "src", ""}, "/")
	return c.path.Set(strings.TrimPrefix(pwd, gopath))
}

// Validate applies custom validation logic to flags and returns an error if any
// flags violate these rules.
func (c *CommandLineFlags) Validate() error {
	if len(c.specs) == 0 {
		return fmt.Errorf("%q flag must not be empty", specFlag)
	}
	if err := c.detectPath(); err != nil {
		return err
	}
	if len(c.path.String()) == 0 {
		return fmt.Errorf("%q flag must not be empty", pathFlag)
	}
	if strings.Contains(c.destination, "..") {
		return fmt.Errorf("destination with '..' in path is not supported")
	}
	if !strings.HasPrefix(c.destination, "."+string(os.PathSeparator)) && c.destination != "." {
		return fmt.Errorf("destination directory must be a relative path")
	}
	return nil
}

// ReadSpecs returns the JSONLD contents of files specified in the 'spec' flag.
func (c *CommandLineFlags) ReadSpecs() (j []rdf.JSONLD, err error) {
	j = make([]rdf.JSONLD, 0, len(c.specs))
	for _, spec := range c.specs {
		var b []byte
		b, err = ioutil.ReadFile(spec)
		if err != nil {
			return
		}
		var inputJSON map[string]interface{}
		err = json.Unmarshal(b, &inputJSON)
		if err != nil {
			return
		}
		j = append(j, inputJSON)
	}
	return
}

// CreateDestination creates the destination path
func (c *CommandLineFlags) CreateDestination() error {
	return os.MkdirAll(c.destination, 0777)
}

// AutoDetectedPath returns true if the path flag was auto-detected.
func (c *CommandLineFlags) AutoDetectedPath() bool {
	return c.pathAutoDetected
}

// Path returns the path flag.
func (c *CommandLineFlags) Path() string {
	return c.path.String()
}

// NewPackageManager creates the correct package manager for the flag inputs.
func (c *CommandLineFlags) NewPackageManager() *gen.PackageManager {
	g := gen.NewPackageManager(c.Path(), "")
	subdirs := strings.Split(
		// Trim "./" prefix as well as "trim" (aka remove) the sole "."
		// path.
		strings.TrimPrefix(
			// Trim "." first
			strings.TrimPrefix(c.destination, "."),
			// Then trim "/"
			string(os.PathSeparator)),
		string(os.PathSeparator))
	for _, subdir := range subdirs {
		g = g.Sub(subdir)
	}
	return g
}

func main() {
	// Read, Parse, and Validate command line flags
	cmd, err := NewCommandLineFlags()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print auto-determined values
	if cmd.AutoDetectedPath() {
		fmt.Printf("Auto-detected path: %s\n", cmd.Path())
	}

	// Create the destination directory
	if err := cmd.CreateDestination(); err != nil {
		fmt.Println(err)
		return
	}

	// Read input specification files
	fmt.Printf("Reading input specifications...\n")
	inputJSONs, err := cmd.ReadSpecs()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Parse specifications
	fmt.Printf("Parsing %d vocabularies...\n", len(inputJSONs))
	p, err := rdf.ParseVocabularies(registry, inputJSONs)
	if err != nil {
		panic(err)
	}

	// Convert to generated code
	fmt.Printf("Converting %d types, properties, and values...\n", p.Size())
	c := &convert.Converter{
		GenRoot:       cmd.NewPackageManager(),
		PackagePolicy: convert.IndividualUnderRoot,
	}
	f, err := c.Convert(p)
	if err != nil {
		panic(err)
	}

	// Write generated code
	fmt.Printf("Writing %d files...\n", len(f))
	for _, file := range f {
		dir := file.Directory
		// If the cwd ("." or "./") are specified as the
		// destination, then the directory may be empty. The cwd does
		// not need to have MkdirAll called on it.
		if dir == "" {
			dir = "."
		} else if e := os.MkdirAll(dir, 0777); e != nil {
			panic(e)
		}

		// Standard generated Go code header.
		// https://github.com/golang/go/issues/13560#issuecomment-288457920
		file.F.HeaderComment("// Code generated by astool. DO NOT EDIT.\n")

		if e := file.F.Save(dir + string(os.PathSeparator) + file.FileName); e != nil {
			panic(e)
		}
	}
	fmt.Printf("Done!\n")
}
