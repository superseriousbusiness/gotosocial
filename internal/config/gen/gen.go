// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/config"
)

const license = `// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

`

func main() {
	var out string

	// Load runtime config flags
	flag.StringVar(&out, "out", "", "Generated file output path")
	flag.Parse()

	// Open output file path
	output, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}

	fmt.Fprint(output, "// THIS IS A GENERATED FILE, DO NOT EDIT BY HAND\n")
	fmt.Fprint(output, license)
	fmt.Fprint(output, "package config\n\n")
	fmt.Fprint(output, "import (\n")
	fmt.Fprint(output, "\t\"time\"\n\n")
	fmt.Fprint(output, "\t\"codeberg.org/gruf/go-bytesize\"\n")
	fmt.Fprint(output, "\t\"github.com/superseriousbusiness/gotosocial/internal/language\"\n")
	fmt.Fprint(output, ")\n\n")
	generateFields(output, nil, reflect.TypeOf(config.Configuration{}))
	_ = output.Close()
	_ = exec.Command("gofumpt", "-w", out).Run()

	// The plan here is that eventually we might be able
	// to generate an example configuration from struct tags
}

func generateFields(output io.Writer, prefixes []string, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if ft := field.Type; ft.Kind() == reflect.Struct {
			// This is a struct field containing further nested config vars.
			generateFields(output, append(prefixes, field.Name), ft)
			continue
		}

		// Get prefixed config variable name
		name := strings.Join(prefixes, "") + field.Name

		// Get period-separated (if nested) config variable "path"
		fieldPath := strings.Join(append(prefixes, field.Name), ".")

		// Get dash-separated config variable CLI flag "path"
		flagPath := strings.Join(append(prefixes, field.Tag.Get("name")), "-")
		flagPath = strings.ToLower(flagPath)

		// Get type without "config." prefix.
		fieldType := strings.ReplaceAll(
			field.Type.String(),
			"config.", "",
		)

		// ConfigState structure helper methods
		fmt.Fprintf(output, "// Get%s safely fetches the Configuration value for state's '%s' field\n", name, fieldPath)
		fmt.Fprintf(output, "func (st *ConfigState) Get%s() (v %s) {\n", name, fieldType)
		fmt.Fprintf(output, "\tst.mutex.RLock()\n")
		fmt.Fprintf(output, "\tv = st.config.%s\n", fieldPath)
		fmt.Fprintf(output, "\tst.mutex.RUnlock()\n")
		fmt.Fprintf(output, "\treturn\n")
		fmt.Fprintf(output, "}\n\n")
		fmt.Fprintf(output, "// Set%s safely sets the Configuration value for state's '%s' field\n", name, fieldPath)
		fmt.Fprintf(output, "func (st *ConfigState) Set%s(v %s) {\n", name, fieldType)
		fmt.Fprintf(output, "\tst.mutex.Lock()\n")
		fmt.Fprintf(output, "\tdefer st.mutex.Unlock()\n")
		fmt.Fprintf(output, "\tst.config.%s = v\n", fieldPath)
		fmt.Fprintf(output, "\tst.reloadToViper()\n")
		fmt.Fprintf(output, "}\n\n")

		// Global ConfigState helper methods
		// TODO: remove when we pass around a ConfigState{}
		fmt.Fprintf(output, "// %sFlag returns the flag name for the '%s' field\n", name, fieldPath)
		fmt.Fprintf(output, "func %sFlag() string { return \"%s\" }\n\n", name, flagPath)
		fmt.Fprintf(output, "// Get%s safely fetches the value for global configuration '%s' field\n", name, fieldPath)
		fmt.Fprintf(output, "func Get%[1]s() %[2]s { return global.Get%[1]s() }\n\n", name, fieldType)
		fmt.Fprintf(output, "// Set%s safely sets the value for global configuration '%s' field\n", name, fieldPath)
		fmt.Fprintf(output, "func Set%[1]s(v %[2]s) { global.Set%[1]s(v) }\n\n", name, fieldType)
	}
}
