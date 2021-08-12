package vm_escaped_indent

import (
	"fmt"

	"github.com/goccy/go-json/internal/encoder"
)

func DebugRun(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet, opt encoder.Option) ([]byte, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("=============[DEBUG]===============")
			fmt.Println("* [TYPE]")
			fmt.Println(codeSet.Type)
			fmt.Printf("\n")
			fmt.Println("* [ALL OPCODE]")
			fmt.Println(codeSet.Code.Dump())
			fmt.Printf("\n")
			fmt.Println("* [CONTEXT]")
			fmt.Printf("%+v\n", ctx)
			fmt.Println("===================================")
			panic(err)
		}
	}()

	return Run(ctx, b, codeSet, opt)
}
