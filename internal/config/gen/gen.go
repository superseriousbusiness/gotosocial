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
		fmt.Fprint(output, "// THIS IS A GENERATED FILE, DO NOT EDIT BY HAND\n")
		fmt.Fprint(output, "package config\n\n")
		t := reflect.TypeOf(config.Configuration{})
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fmt.Fprintf(output, "// %sFlag returns the flag name for the '%s' field\n", field.Name, field.Name)
			fmt.Fprintf(output, "func %sFlag() string {\n", field.Name)
			fmt.Fprintf(output, "\treturn \"%s\"\n", field.Tag.Get("name"))
			fmt.Fprintf(output, "}\n\n")
			fmt.Fprintf(output, "// Get%s safely fetches the value for global configuration '%s' field\n", field.Name, field.Name)
			fmt.Fprintf(output, "func Get%s() (v %s) {\n", field.Name, field.Type.String())
			fmt.Fprintf(output, "\tmutex.Lock()\n")
			fmt.Fprintf(output, "\tv = global.%s\n", field.Name)
			fmt.Fprintf(output, "\tmutex.Unlock()\n")
			fmt.Fprintf(output, "\treturn\n")
			fmt.Fprintf(output, "}\n\n")
			fmt.Fprintf(output, "// Set%s safely sets the value for global configuration '%s' field\n", field.Name, field.Name)
			fmt.Fprintf(output, "func Set%s(v %s) {\n", field.Name, field.Type.String())
			fmt.Fprintf(output, "\tmutex.Lock()\n")
			fmt.Fprintf(output, "\tglobal.%s = v\n", field.Name)
			fmt.Fprintf(output, "\tmutex.Unlock()\n")
			fmt.Fprintf(output, "}\n\n")
		}
		output.Close()
		_ = exec.Command("gofmt", "-w", out).Run()

	// Unknown type
	default:
		panic("unknown generation type: " + gen)
	}
}
