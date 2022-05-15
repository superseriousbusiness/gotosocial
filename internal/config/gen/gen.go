package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/superseriousbusiness/gotosocial/internal/config"
)

func main() {
	var (
		out string
		gen string
	)

	// Load runtime config flags
	flag.StringVar(&out, "out", "", "Generated file output path")
	flag.StringVar(&gen, "gen", "values", "Type of file to generate (values)")
	flag.Parse()

	// Open output file path
	output, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}

	switch gen {
	// Generate config field helper methods
	case "values":
		fmt.Fprint(output, "package config\n\n")
		t := reflect.TypeOf(config.Configuration{})
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fmt.Fprintf(output, "// Get%s safely fetches the value for global configuration '%s' field\n", field.Name, field.Name)
			fmt.Fprintf(output, "func Get%s() (v %s) {\n", field.Name, field.Type.String())
			fmt.Fprintf(output, "\tConfig(func(cfg *Configuration) {\n")
			fmt.Fprintf(output, "\t\tv = cfg.%s\n", field.Name)
			fmt.Fprintf(output, "\t})\n")
			fmt.Fprintf(output, "\treturn\n")
			fmt.Fprintf(output, "}\n\n")
		}
		output.Close()
		exec.Command("gofmt", "-w", out).Run()

	// Unknown type
	default:
		panic("unknown generation type: " + gen)
	}
}
