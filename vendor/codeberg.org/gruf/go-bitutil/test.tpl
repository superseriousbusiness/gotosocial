package atomics_test

import (
    "atomic"
    "unsafe"
    "testing"

    "codeberg.org/gruf/go-atomics"
)

func Test{{ .Name }}StoreLoad(t *testing.T) {
    for _, test := range {{ .Name }}Tests {
        val := atomics.New{{ .Name }}()

        val.Store(test.V1)

        if !({{ call .Compare "val.Load()" "test.V1" }}) {
            t.Fatalf("failed testing .Store and .Load: expect=%v actual=%v", val.Load(), test.V1)
        }

        val.Store(test.V2)

        if !({{ call .Compare "val.Load()" "test.V2" }}) {
            t.Fatalf("failed testing .Store and .Load: expect=%v actual=%v", val.Load(), test.V2)
        }
    }
}

func Test{{ .Name }}CAS(t *testing.T) {
    for _, test := range {{ .Name }}Tests {
        val := atomics.New{{ .Name }}()

        val.Store(test.V1)

        if val.CAS(test.V2, test.V1) {
            t.Fatalf("failed testing negative .CAS: test=%+v state=%v", test, val.Load())
        }

        if !val.CAS(test.V1, test.V2) {
            t.Fatalf("failed testing positive .CAS: test=%+v state=%v", test, val.Load())
        }
    }
}

func Test{{ .Name }}Swap(t *testing.T) {
    for _, test := range {{ .Name }}Tests {
        val := atomics.New{{ .Name }}()

        val.Store(test.V1)

        if !({{ call .Compare "val.Swap(test.V2)" "test.V1" }}) {
            t.Fatal("failed testing .Swap")
        }

        if !({{ call .Compare "val.Swap(test.V1)" "test.V2" }}) {
            t.Fatal("failed testing .Swap")
        }
    }
}

