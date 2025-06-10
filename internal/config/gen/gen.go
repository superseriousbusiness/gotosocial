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
	"slices"
	"strings"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/config"
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

var durationType = reflect.TypeOf(time.Duration(0))
var stringerType = reflect.TypeOf((*interface{ String() string })(nil)).Elem()
var stringersType = reflect.TypeOf((*interface{ Strings() []string })(nil)).Elem()
var flagSetType = reflect.TypeOf((*interface{ Set(string) error })(nil)).Elem()

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

	configType := reflect.TypeOf(config.Configuration{})

	// Parse our config type for usable fields.
	fields := loadConfigFields(nil, nil, configType)

	fprintf(output, "// THIS IS A GENERATED FILE, DO NOT EDIT BY HAND\n")
	fprintf(output, license)
	fprintf(output, "package config\n\n")
	fprintf(output, "import (\n")
	fprintf(output, "\t\"fmt\"\n")
	fprintf(output, "\t\"time\"\n\n")
	fprintf(output, "\t\"codeberg.org/gruf/go-bytesize\"\n")
	fprintf(output, "\t\"code.superseriousbusiness.org/gotosocial/internal/language\"\n")
	fprintf(output, "\t\"github.com/spf13/pflag\"\n")
	fprintf(output, "\t\"github.com/spf13/cast\"\n")
	fprintf(output, ")\n")
	fprintf(output, "\n")
	generateFlagConsts(output, fields)
	generateFlagRegistering(output, fields)
	generateMapMarshaler(output, fields)
	generateMapUnmarshaler(output, fields)
	generateGetSetters(output, fields)
	generateMapFlattener(output, fields)
	must(output.Close())
	must(exec.Command("gofumpt", "-w", out).Run())
}

type ConfigField struct {
	// Any CLI flag prefixes,
	// i.e. with nested fields.
	Prefixes []string

	// The base CLI flag
	// name of the field.
	Name string

	// Path to struct field
	// in dot-separated form.
	Path string

	// Usage string.
	Usage string

	// The underlying Go type
	// of the config field.
	Type reflect.Type

	// i.e. is this found in the configuration file?
	// or just used in specific CLI commands? in the
	// future we'll remove these from config struct.
	Ephemeral bool
}

// Flag returns the combined "prefixes-name" CLI flag for config field.
func (f ConfigField) Flag() string {
	flag := strings.Join(append(f.Prefixes, f.Name), "-")
	flag = strings.ToLower(flag)
	return flag
}

// PossibleKeys returns a list of possible map key combinations
// that this config field may be found under. The combined "prefixes-name"
// will always be in the list, but also separates them out to account for
// possible nesting. This allows us to support both nested and un-nested
// configuration files, always prioritizing "prefixes-name" as its the CLI flag.
func (f ConfigField) PossibleKeys() [][]string {
	if len(f.Prefixes) == 0 {
		return [][]string{{f.Name}}
	}

	var keys [][]string

	combined := f.Flag()
	keys = append(keys, []string{combined})

	basePrefix := strings.TrimSuffix(combined, "-"+f.Name)
	keys = append(keys, []string{basePrefix, f.Name})

	for i := len(f.Prefixes) - 1; i >= 0; i-- {
		prefix := f.Prefixes[i]

		basePrefix = strings.TrimSuffix(basePrefix, prefix)
		basePrefix = strings.TrimSuffix(basePrefix, "-")
		if len(basePrefix) == 0 {
			break
		}

		var key []string
		key = append(key, basePrefix)
		key = append(key, f.Prefixes[i:]...)
		key = append(key, f.Name)
		keys = append(keys, key)
	}

	return keys
}

func loadConfigFields(pathPrefixes, flagPrefixes []string, t reflect.Type) []ConfigField {
	var out []ConfigField
	for i := 0; i < t.NumField(); i++ {
		// Struct field at index.
		field := t.Field(i)

		// Get field's tagged name.
		name := field.Tag.Get("name")
		if name == "" || name == "-" {
			continue
		}

		if ft := field.Type; ft.Kind() == reflect.Struct {
			// This is a nested struct, load nested fields.
			pathPrefixes := append(pathPrefixes, field.Name)
			flagPrefixes := append(flagPrefixes, name)
			out = append(out, loadConfigFields(pathPrefixes, flagPrefixes, ft)...)
			continue
		}

		// Get prefixed, period-separated, config variable struct "path".
		fieldPath := strings.Join(append(pathPrefixes, field.Name), ".")

		// Append prepared ConfigField.
		out = append(out, ConfigField{
			Prefixes:  flagPrefixes,
			Name:      name,
			Path:      fieldPath,
			Usage:     field.Tag.Get("usage"),
			Ephemeral: field.Tag.Get("ephemeral") == "yes",
			Type:      field.Type,
		})
	}
	return out
}

func generateFlagConsts(out io.Writer, fields []ConfigField) {
	fprintf(out, "const (\n")
	for _, field := range fields {
		name := strings.ReplaceAll(field.Path, ".", "")
		fprintf(out, "\t%sFlag = \"%s\"\n", name, field.Flag())
	}
	fprintf(out, ")\n\n")
}

func generateFlagRegistering(out io.Writer, fields []ConfigField) {
	fprintf(out, "func (cfg *Configuration) RegisterFlags(flags *pflag.FlagSet) {\n")
	for _, field := range fields {
		if field.Ephemeral {
			// Skip registering
			// ephemeral flags.
			continue
		}

		// Check for easy cases of just regular primitive types.
		if field.Type.Kind().String() == field.Type.String() {
			typeName := field.Type.String()
			typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
			fprintf(out, "\tflags.%s(\"%s\", cfg.%s, \"%s\")\n", typeName, field.Flag(), field.Path, field.Usage)
			continue
		}

		// Check for easy cases of just
		// regular primitive slice types.
		if field.Type.Kind() == reflect.Slice {
			elem := field.Type.Elem()
			if elem.Kind().String() == elem.String() {
				typeName := elem.String()
				typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
				fprintf(out, "\tflags.%sSlice(\"%s\", cfg.%s, \"%s\")\n", typeName, field.Flag(), field.Path, field.Usage)
				continue
			}
		}

		// Durations should get set directly
		// as their types as viper knows how
		// to deal with this type directly.
		if field.Type == durationType {
			fprintf(out, "\tflags.Duration(\"%s\", cfg.%s, \"%s\")\n", field.Flag(), field.Path, field.Usage)
			continue
		}

		if field.Type.Kind() == reflect.Slice {
			// Check if the field supports Stringers{}.
			if field.Type.Implements(stringersType) {
				fprintf(out, "\tflags.StringSlice(\"%s\", cfg.%s.Strings(), \"%s\")\n", field.Flag(), field.Path, field.Usage)
				continue
			}

			// Or the pointer type of the field value supports Stringers{}.
			if ptr := reflect.PointerTo(field.Type); ptr.Implements(stringersType) {
				fprintf(out, "\tflags.StringSlice(\"%s\", cfg.%s.Strings(), \"%s\")\n", field.Flag(), field.Path, field.Usage)
				continue
			}

			fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Path, stringersType)
		} else {
			// Check if the field supports Stringer{}.
			if field.Type.Implements(stringerType) {
				fprintf(out, "\tflags.String(\"%s\", cfg.%s.String(), \"%s\")\n", field.Flag(), field.Path, field.Usage)
				continue
			}

			// Or the pointer type of the field value supports Stringer{}.
			if ptr := reflect.PointerTo(field.Type); ptr.Implements(stringerType) {
				fprintf(out, "\tflags.String(\"%s\", cfg.%s.String(), \"%s\")\n", field.Flag(), field.Path, field.Usage)
				continue
			}

			fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Path, stringerType)
		}
	}
	fprintf(out, "}\n\n")
}

func generateMapMarshaler(out io.Writer, fields []ConfigField) {
	fprintf(out, "func (cfg *Configuration) MarshalMap() map[string]any {\n")
	fprintf(out, "\tcfgmap := make(map[string]any, %d)\n", len(fields))
	for _, field := range fields {
		// Check for easy cases of just regular primitive types.
		if field.Type.Kind().String() == field.Type.String() {
			fprintf(out, "\tcfgmap[\"%s\"] = cfg.%s\n", field.Flag(), field.Path)
			continue
		}

		// Check for easy cases of just
		// regular primitive slice types.
		if field.Type.Kind() == reflect.Slice {
			elem := field.Type.Elem()
			if elem.Kind().String() == elem.String() {
				fprintf(out, "\tcfgmap[\"%s\"] = cfg.%s\n", field.Flag(), field.Path)
				continue
			}
		}

		// Durations should get set directly
		// as their types as viper knows how
		// to deal with this type directly.
		if field.Type == durationType {
			fprintf(out, "\tcfgmap[\"%s\"] = cfg.%s\n", field.Flag(), field.Path)
			continue
		}

		if field.Type.Kind() == reflect.Slice {
			// Either the field must support Stringers{}.
			if field.Type.Implements(stringersType) {
				fprintf(out, "\tcfgmap[\"%s\"] = cfg.%s.Strings()\n", field.Flag(), field.Path)
				continue
			}

			// Or the pointer type of the field value must support Stringers{}.
			if ptr := reflect.PointerTo(field.Type); ptr.Implements(stringersType) {
				fprintf(out, "\tcfgmap[\"%s\"] = cfg.%s.Strings()\n", field.Flag(), field.Path)
				continue
			}

			fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Path, stringersType)
		} else {
			// Either the field must support Stringer{}.
			if field.Type.Implements(stringerType) {
				fprintf(out, "\tcfgmap[\"%s\"] = cfg.%s.String()\n", field.Flag(), field.Path)
				continue
			}

			// Or the pointer type of the field value must support Stringer{}.
			if ptr := reflect.PointerTo(field.Type); ptr.Implements(stringerType) {
				fprintf(out, "\tcfgmap[\"%s\"] = cfg.%s.String()\n", field.Flag(), field.Path)
				continue
			}

			fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Path, stringerType)
		}
	}
	fprintf(out, "\treturn cfgmap")
	fprintf(out, "}\n\n")
}

func generateMapUnmarshaler(out io.Writer, fields []ConfigField) {
	fprintf(out, "func (cfg *Configuration) UnmarshalMap(cfgmap map[string]any) error {\n")
	fprintf(out, "// VERY IMPORTANT FIRST STEP!\n")
	fprintf(out, "// flatten to normalize map to\n")
	fprintf(out, "// entirely un-nested key values\n")
	fprintf(out, "flattenConfigMap(cfgmap)\n")
	fprintf(out, "\n")
	for _, field := range fields {
		// Check for easy cases of just regular primitive types.
		if field.Type.Kind().String() == field.Type.String() {
			generateUnmarshalerPrimitive(out, field)
			continue
		}

		// Check for easy cases of just
		// regular primitive slice types.
		if field.Type.Kind() == reflect.Slice {
			elem := field.Type.Elem()
			if elem.Kind().String() == elem.String() {
				generateUnmarshalerPrimitive(out, field)
				continue
			}
		}

		// Durations should get set directly
		// as their types as viper knows how
		// to deal with this type directly.
		if field.Type == durationType {
			generateUnmarshalerPrimitive(out, field)
			continue
		}

		// Either the field must support flag.Value{}.
		if field.Type.Implements(flagSetType) {
			generateUnmarshalerFlagType(out, field)
			continue
		}

		// Or the pointer type of the field value must support flag.Value{}.
		if ptr := reflect.PointerTo(field.Type); ptr.Implements(flagSetType) {
			generateUnmarshalerFlagType(out, field)
			continue
		}

		fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Path, flagSetType)
	}
	fprintf(out, "\treturn nil\n")
	fprintf(out, "}\n\n")
}

func generateUnmarshalerPrimitive(out io.Writer, field ConfigField) {
	fprintf(out, "\t\tif ival, ok := cfgmap[\"%s\"]; ok {\n", field.Flag())
	if field.Type.Kind() == reflect.Slice {
		elem := field.Type.Elem()
		typeName := elem.String()
		if i := strings.IndexRune(typeName, '.'); i >= 0 {
			typeName = typeName[i+1:]
		}
		typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
		fprintf(out, "\t\t\tvar err error\n")
		// note we specifically handle slice types ourselves to split by comma
		fprintf(out, "\t\t\tcfg.%s, err = to%sSlice(ival)\n", field.Path, typeName)
		fprintf(out, "\t\t\tif err != nil {\n")
		fprintf(out, "\t\t\t\treturn fmt.Errorf(\"error casting %%#v -> []%s for '%s': %%w\", ival, err)\n", elem.String(), field.Flag())
		fprintf(out, "\t\t\t}\n")
	} else {
		typeName := field.Type.String()
		if i := strings.IndexRune(typeName, '.'); i >= 0 {
			typeName = typeName[i+1:]
		}
		typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
		fprintf(out, "\t\t\tvar err error\n")
		fprintf(out, "\t\t\tcfg.%s, err = cast.To%sE(ival)\n", field.Path, typeName)
		fprintf(out, "\t\t\tif err != nil {\n")
		fprintf(out, "\t\t\t\treturn fmt.Errorf(\"error casting %%#v -> %s for '%s': %%w\", ival, err)\n", field.Type.String(), field.Flag())
		fprintf(out, "\t\t\t}\n")
	}
	fprintf(out, "\t}\n")
	fprintf(out, "\n")
}

func generateUnmarshalerFlagType(out io.Writer, field ConfigField) {
	fprintf(out, "\t\tif ival, ok := cfgmap[\"%s\"]; ok {\n", field.Flag())
	if field.Type.Kind() == reflect.Slice {
		// same as above re: slice types and splitting on comma
		fprintf(out, "\t\tt, err := toStringSlice(ival)\n")
		fprintf(out, "\t\tif err != nil {\n")
		fprintf(out, "\t\t\treturn fmt.Errorf(\"error casting %%#v -> []string for '%s': %%w\", ival, err)\n", field.Flag())
		fprintf(out, "\t\t}\n")
		fprintf(out, "\t\tcfg.%s = %s{}\n", field.Path, strings.TrimPrefix(field.Type.String(), "config."))
		fprintf(out, "\t\tfor _, in := range t {\n")
		fprintf(out, "\t\t\tif err := cfg.%s.Set(in); err != nil {\n", field.Path)
		fprintf(out, "\t\t\t\treturn fmt.Errorf(\"error parsing %%#v for '%s': %%w\", ival, err)\n", field.Flag())
		fprintf(out, "\t\t\t}\n")
		fprintf(out, "\t\t}\n")
	} else {
		fprintf(out, "\t\tt, err := cast.ToStringE(ival)\n")
		fprintf(out, "\t\tif err != nil {\n")
		fprintf(out, "\t\t\treturn fmt.Errorf(\"error casting %%#v -> string for '%s': %%w\", ival, err)\n", field.Flag())
		fprintf(out, "\t\t}\n")
		fprintf(out, "\t\tcfg.%s = %#v\n", field.Path, reflect.New(field.Type).Elem().Interface())
		fprintf(out, "\t\tif err := cfg.%s.Set(t); err != nil {\n", field.Path)
		fprintf(out, "\t\t\treturn fmt.Errorf(\"error parsing %%#v for '%s': %%w\", ival, err)\n", field.Flag())
		fprintf(out, "\t\t}\n")
	}
	fprintf(out, "\t}\n")
	fprintf(out, "\n")
}

func generateGetSetters(out io.Writer, fields []ConfigField) {
	for _, field := range fields {
		// Get name from struct path, without periods.
		name := strings.ReplaceAll(field.Path, ".", "")

		// Get type without "config." prefix.
		fieldType := strings.ReplaceAll(
			field.Type.String(),
			"config.", "",
		)

		// ConfigState structure helper methods
		fprintf(out, "// Get%s safely fetches the Configuration value for state's '%s' field\n", name, field.Path)
		fprintf(out, "func (st *ConfigState) Get%s() (v %s) {\n", name, fieldType)
		fprintf(out, "\tst.mutex.RLock()\n")
		fprintf(out, "\tv = st.config.%s\n", field.Path)
		fprintf(out, "\tst.mutex.RUnlock()\n")
		fprintf(out, "\treturn\n")
		fprintf(out, "}\n\n")
		fprintf(out, "// Set%s safely sets the Configuration value for state's '%s' field\n", name, field.Path)
		fprintf(out, "func (st *ConfigState) Set%s(v %s) {\n", name, fieldType)
		fprintf(out, "\tst.mutex.Lock()\n")
		fprintf(out, "\tdefer st.mutex.Unlock()\n")
		fprintf(out, "\tst.config.%s = v\n", field.Path)
		fprintf(out, "\tst.reloadToViper()\n")
		fprintf(out, "}\n\n")

		// Global ConfigState helper methods
		fprintf(out, "// Get%s safely fetches the value for global configuration '%s' field\n", name, field.Path)
		fprintf(out, "func Get%[1]s() %[2]s { return global.Get%[1]s() }\n\n", name, fieldType)
		fprintf(out, "// Set%s safely sets the value for global configuration '%s' field\n", name, field.Path)
		fprintf(out, "func Set%[1]s(v %[2]s) { global.Set%[1]s(v) }\n\n", name, fieldType)
	}

	// Separate out the config fields (from a clone!!!) to get only the 'mem-ratio' members.
	memFields := slices.DeleteFunc(slices.Clone(fields), func(field ConfigField) bool {
		return !strings.Contains(field.Path, "MemRatio")
	})

	fprintf(out, "// GetTotalOfMemRatios safely fetches the combined value for all the state's mem ratio fields\n")
	fprintf(out, "func (st *ConfigState) GetTotalOfMemRatios() (total float64) {\n")
	fprintf(out, "\tst.mutex.RLock()\n")
	for _, field := range memFields {
		fprintf(out, "\ttotal += st.config.%s\n", field.Path)
	}
	fprintf(out, "\tst.mutex.RUnlock()\n")
	fprintf(out, "\treturn\n")
	fprintf(out, "}\n\n")

	fprintf(out, "// GetTotalOfMemRatios safely fetches the combined value for all the global state's mem ratio fields\n")
	fprintf(out, "func GetTotalOfMemRatios() (total float64) { return global.GetTotalOfMemRatios() }\n\n")
}

func generateMapFlattener(out io.Writer, fields []ConfigField) {
	fprintf(out, "func flattenConfigMap(cfgmap map[string]any) {\n")
	fprintf(out, "\tnestedKeys := make(map[string]struct{})\n")
	for _, field := range fields {
		keys := field.PossibleKeys()
		if len(keys) <= 1 {
			continue
		}
		fprintf(out, "\tfor _, key := range [][]string{\n")
		for _, key := range keys[1:] {
			fprintf(out, "\t\t{\"%s\"},\n", strings.Join(key, "\", \""))
		}
		fprintf(out, "\t} {\n")
		fprintf(out, "\t\tival, ok := mapGet(cfgmap, key...)\n")
		fprintf(out, "\t\tif ok {\n")
		fprintf(out, "\t\t\tcfgmap[\"%s\"] = ival\n", field.Flag())
		fprintf(out, "\t\t\tnestedKeys[key[0]] = struct{}{}\n")
		fprintf(out, "\t\t\tbreak\n")
		fprintf(out, "\t\t}\n")
		fprintf(out, "\t}\n\n")
	}
	fprintf(out, "\tfor key := range nestedKeys {\n")
	fprintf(out, "\t\tdelete(cfgmap, key)\n")
	fprintf(out, "\t}\n")
	fprintf(out, "}\n\n")
}

func fprintf(out io.Writer, format string, args ...any) {
	_, err := fmt.Fprintf(out, format, args...)
	must(err)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
