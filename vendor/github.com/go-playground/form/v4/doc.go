/*
Package form Decodes url.Values into Go value(s) and Encodes Go value(s) into url.Values.


It has the following features:

    - Primitives types cause zero allocations.
    - Supports map of almost all types.
    - Supports both Numbered and Normal arrays eg. "Array[0]" and just "Array"
      with multiple values passed.
    - Slice honours the specified index. eg. if "Slice[2]" is the only Slice
      value passed down, it will be put at index 2; if slice isn't big enough
      it will be expanded.
    - Array honours the specified index. eg. if "Array[2]" is the only Array
      value passed down, it will be put at index 2; if array isn't big enough
      a warning will be printed and value ignored.
    - Only creates objects as necessary eg. if no `array` or `map` values are
      passed down, the `array` and `map` are left as their default values in
      the struct.
    - Allows for Custom Type registration.
    - Handles time.Time using RFC3339 time format by default,
      but can easily be changed by registering a Custom Type, see below.
    - Handles Encoding & Decoding of almost all Go types eg. can Decode into
      struct, array, map, int... and Encode a struct, array, map, int...

Common Questions

Questions

    Does it support encoding.TextUnmarshaler?
    No because TextUnmarshaler only accepts []byte but posted values can have
    multiple values, so is not suitable.

	Mixing array/slice with array[idx]/slice[idx], in which order are they parsed?
	array/slice then array[idx]/slice[idx]

Supported Types

out of the box supported types

    - string
    - bool
    - int, int8, int16, int32, int64
    - uint, uint8, uint16, uint32, uint64
    - float32, float64
    - struct and anonymous struct
    - interface{}
    - time.Time` - by default using RFC3339
    - a `pointer` to one of the above types
    - slice, array
    - map
    - `custom types` can override any of the above types
    - many other types may be supported inherently (eg. bson.ObjectId is
      type ObjectId string, which will get populated by the string type

    **NOTE**: map, struct and slice nesting are ad infinitum.

Usage

symbols

    - Use symbol `.` for separating fields/structs. (eg. `structfield.field`)
    - Use `[index or key]` for access to index of a slice/array or key for map.
      (eg. `arrayfield[0]`, `mapfield[keyvalue]`)

html

    <form method="POST">
        <input type="text" name="Name" value="joeybloggs"/>
        <input type="text" name="Age" value="3"/>
        <input type="text" name="Gender" value="Male"/>
        <input type="text" name="Address[0].Name" value="26 Here Blvd."/>
        <input type="text" name="Address[0].Phone" value="9(999)999-9999"/>
        <input type="text" name="Address[1].Name" value="26 There Blvd."/>
        <input type="text" name="Address[1].Phone" value="1(111)111-1111"/>
        <input type="text" name="active" value="true"/>
        <input type="text" name="MapExample[key]" value="value"/>
        <input type="text" name="NestedMap[key][key]" value="value"/>
        <input type="text" name="NestedArray[0][0]" value="value"/>
        <input type="submit"/>
    </form>

Example

example decoding the above HTML

    package main

    import (
        "fmt"
        "log"
        "net/url"

        "github.com/go-playground/form/v4"
    )

    // Address contains address information
    type Address struct {
        Name  string
        Phone string
    }

    // User contains user information
    type User struct {
        Name        string
        Age         uint8
        Gender      string
        Address     []Address
        Active      bool `form:"active"`
        MapExample  map[string]string
        NestedMap   map[string]map[string]string
        NestedArray [][]string
    }

    // use a single instance of Decoder, it caches struct info
    var decoder *form.Decoder

    func main() {
        decoder = form.NewDecoder()

        // this simulates the results of http.Request's ParseForm() function
        values := parseForm()

        var user User

        // must pass a pointer
        err := decoder.Decode(&user, values)
        if err != nil {
            log.Panic(err)
        }

        fmt.Printf("%#v\n", user)
    }

    // this simulates the results of http.Request's ParseForm() function
    func parseForm() url.Values {
        return url.Values{
            "Name":                []string{"joeybloggs"},
            "Age":                 []string{"3"},
            "Gender":              []string{"Male"},
            "Address[0].Name":     []string{"26 Here Blvd."},
            "Address[0].Phone":    []string{"9(999)999-9999"},
            "Address[1].Name":     []string{"26 There Blvd."},
            "Address[1].Phone":    []string{"1(111)111-1111"},
            "active":              []string{"true"},
            "MapExample[key]":     []string{"value"},
            "NestedMap[key][key]": []string{"value"},
            "NestedArray[0][0]":   []string{"value"},
        }
    }

example encoding

    package main

    import (
        "fmt"
        "log"

        "github.com/go-playground/form/v4"
    )

    // Address contains address information
    type Address struct {
        Name  string
        Phone string
    }

    // User contains user information
    type User struct {
        Name        string
        Age         uint8
        Gender      string
        Address     []Address
        Active      bool `form:"active"`
        MapExample  map[string]string
        NestedMap   map[string]map[string]string
        NestedArray [][]string
    }

    // use a single instance of Encoder, it caches struct info
    var encoder *form.Encoder

    func main() {
        encoder = form.NewEncoder()

        user := User{
            Name:   "joeybloggs",
            Age:    3,
            Gender: "Male",
            Address: []Address{
                {Name: "26 Here Blvd.", Phone: "9(999)999-9999"},
                {Name: "26 There Blvd.", Phone: "1(111)111-1111"},
            },
            Active:      true,
            MapExample:  map[string]string{"key": "value"},
            NestedMap:   map[string]map[string]string{"key": {"key": "value"}},
            NestedArray: [][]string{{"value"}},
        }

        // must pass a pointer
        values, err := encoder.Encode(&user)
        if err != nil {
            log.Panic(err)
        }

        fmt.Printf("%#v\n", values)
    }


Registering Custom Types

Decoder

    decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
            return time.Parse("2006-01-02", vals[0])
        }, time.Time{})

    ADDITIONAL: if a struct type is registered, the function will only be called
    if a url.Value exists for the struct and not just the struct fields
    eg. url.Values{"User":"Name%3Djoeybloggs"} will call the custom type function
    with 'User' as the type, however url.Values{"User.Name":"joeybloggs"} will not.

Encoder

    encoder.RegisterCustomTypeFunc(func(x interface{}) ([]string, error) {
            return []string{x.(time.Time).Format("2006-01-02")}, nil
        }, time.Time{})


Ignoring Fields

you can tell form to ignore fields using `-` in the tag

    type MyStruct struct {
        Field string `form:"-"`
    }

Omitempty

you can tell form to omit empty fields using `,omitempty` or `FieldName,omitempty` in the tag

    type MyStruct struct {
        Field  string `form:",omitempty"`
        Field2 string `form:"CustomFieldName,omitempty"`
    }


Notes

To maximize compatibility with other systems the Encoder attempts
to avoid using array indexes in url.Values if at all possible.

    eg.

    // A struct field of
    Field []string{"1", "2", "3"}

    // will be output a url.Value as
    "Field": []string{"1", "2", "3"}

    and not
    "Field[0]": []string{"1"}
    "Field[1]": []string{"2"}
    "Field[2]": []string{"3"}

    // however there are times where it is unavoidable, like with pointers
    i := int(1)
    Field []*string{nil, nil, &i}

    // to avoid index 1 and 2 must use index
    "Field[2]": []string{"1"}

*/
package form
