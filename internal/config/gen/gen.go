/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/superseriousbusiness/gotosocial/internal/config"
)

const license = `/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
`

func main() {
	var (
		out string
		gen string
	)

	// Load runtime config flags
	flag.StringVar(&out, "out", "", "Generated file output path")
	flag.StringVar(&gen, "gen", "helpers", "Type of file to generate (helpers)")
	flag.Parse()

	// Open output file path
	output, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}

	switch gen {
	// Generate config field helper methods
	case "helpers":
		fmt.Fprint(output, "// THIS IS A GENERATED FILE, DO NOT EDIT BY HAND\n")
		fmt.Fprint(output, license)
		fmt.Fprint(output, "package config\n\n")
		t := reflect.TypeOf(config.Configuration{})
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// ConfigState structure helper methods
			fmt.Fprintf(output, "// Get%s safely fetches the Configuration value for state's '%s' field\n", field.Name, field.Name)
			fmt.Fprintf(output, "func (st *ConfigState) Get%s() (v %s) {\n", field.Name, field.Type.String())
			fmt.Fprintf(output, "\tst.mutex.Lock()\n")
			fmt.Fprintf(output, "\tv = st.config.%s\n", field.Name)
			fmt.Fprintf(output, "\tst.mutex.Unlock()\n")
			fmt.Fprintf(output, "\treturn\n")
			fmt.Fprintf(output, "}\n\n")
			fmt.Fprintf(output, "// Set%s safely sets the Configuration value for state's '%s' field\n", field.Name, field.Name)
			fmt.Fprintf(output, "func (st *ConfigState) Set%s(v %s) {\n", field.Name, field.Type.String())
			fmt.Fprintf(output, "\tst.mutex.Lock()\n")
			fmt.Fprintf(output, "\tdefer st.mutex.Unlock()\n")
			fmt.Fprintf(output, "\tst.config.%s = v\n", field.Name)
			fmt.Fprintf(output, "\tst.reloadToViper()\n")
			fmt.Fprintf(output, "}\n\n")

			// Global ConfigState helper methods
			// TODO: remove when we pass around a ConfigState{}
			fmt.Fprintf(output, "// %sFlag returns the flag name for the '%s' field\n", field.Name, field.Name)
			fmt.Fprintf(output, "func %sFlag() string { return \"%s\" }\n\n", field.Name, field.Tag.Get("name"))
			fmt.Fprintf(output, "// Get%s safely fetches the value for global configuration '%s' field\n", field.Name, field.Name)
			fmt.Fprintf(output, "func Get%[1]s() %[2]s { return global.Get%[1]s() }\n\n", field.Name, field.Type.String())
			fmt.Fprintf(output, "// Set%s safely sets the value for global configuration '%s' field\n", field.Name, field.Name)
			fmt.Fprintf(output, "func Set%[1]s(v %[2]s) { global.Set%[1]s(v) }\n\n", field.Name, field.Type.String())
		}
		_ = output.Close()
		_ = exec.Command("gofmt", "-w", out).Run()

	// The plain here is that eventually we might be able
	// to generate an example configuration from struct tags

	// Unknown type
	default:
		panic("unknown generation type: " + gen)
	}
}
