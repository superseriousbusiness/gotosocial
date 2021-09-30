// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package v8go

// #include <stdlib.h>
// #include "v8go.h"
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// Due to the limitations of passing pointers to C from Go we need to create
// a registry so that we can lookup the Context from any given callback from V8.
// This is similar to what is described here: https://github.com/golang/go/wiki/cgo#function-variables
// To make sure we can still GC *Context we register the context only when we are
// running a script inside the context and then deregister.
type ctxRef struct {
	ctx      *Context
	refCount int
}

var ctxMutex sync.RWMutex
var ctxRegistry = make(map[int]*ctxRef)
var ctxSeq = 0

// Context is a global root execution environment that allows separate,
// unrelated, JavaScript applications to run in a single instance of V8.
type Context struct {
	ref int
	ptr C.ContextPtr
	iso *Isolate
}

type contextOptions struct {
	iso   *Isolate
	gTmpl *ObjectTemplate
}

// ContextOption sets options such as Isolate and Global Template to the NewContext
type ContextOption interface {
	apply(*contextOptions)
}

// NewContext creates a new JavaScript context; if no Isolate is passed as a
// ContextOption than a new Isolate will be created.
func NewContext(opt ...ContextOption) (*Context, error) {
	opts := contextOptions{}
	for _, o := range opt {
		if o != nil {
			o.apply(&opts)
		}
	}

	if opts.iso == nil {
		var err error
		opts.iso, err = NewIsolate()
		if err != nil {
			return nil, fmt.Errorf("v8go: failed to create new Isolate: %v", err)
		}
	}

	if opts.gTmpl == nil {
		opts.gTmpl = &ObjectTemplate{&template{}}
	}

	ctxMutex.Lock()
	ctxSeq++
	ref := ctxSeq
	ctxMutex.Unlock()

	ctx := &Context{
		ref: ref,
		ptr: C.NewContext(opts.iso.ptr, opts.gTmpl.ptr, C.int(ref)),
		iso: opts.iso,
	}
	// TODO: [RC] catch any C++ exceptions and return as error
	return ctx, nil
}

// Isolate gets the current context's parent isolate.An  error is returned
// if the isolate has been terninated.
func (c *Context) Isolate() (*Isolate, error) {
	// TODO: [RC] check to see if the isolate has not been terninated
	return c.iso, nil
}

// RunScript executes the source JavaScript; origin or filename provides a
// reference for the script and used in the stack trace if there is an error.
// error will be of type `JSError` of not nil.
func (c *Context) RunScript(source string, origin string) (*Value, error) {
	cSource := C.CString(source)
	cOrigin := C.CString(origin)
	defer C.free(unsafe.Pointer(cSource))
	defer C.free(unsafe.Pointer(cOrigin))

	c.register()
	rtn := C.RunScript(c.ptr, cSource, cOrigin)
	c.deregister()

	return getValue(c, rtn), getError(rtn)
}

// Global returns the global proxy object.
// Global proxy object is a thin wrapper whose prototype points to actual
// context's global object with the properties like Object, etc. This is
// done that way for security reasons.
// Please note that changes to global proxy object prototype most probably
// would break the VM — V8 expects only global object as a prototype of
// global proxy object.
func (c *Context) Global() *Object {
	valPtr := C.ContextGlobal(c.ptr)
	v := &Value{valPtr, c}
	return &Object{v}
}

// PerformMicrotaskCheckpoint runs the default MicrotaskQueue until empty.
// This is used to make progress on Promises.
func (c *Context) PerformMicrotaskCheckpoint() {
	c.register()
	defer c.deregister()
	C.IsolatePerformMicrotaskCheckpoint(c.iso.ptr)
}

// Close will dispose the context and free the memory.
// Access to any values assosiated with the context after calling Close may panic.
func (c *Context) Close() {
	C.ContextFree(c.ptr)
	c.ptr = nil
}

func (c *Context) register() {
	ctxMutex.Lock()
	r := ctxRegistry[c.ref]
	if r == nil {
		r = &ctxRef{ctx: c}
		ctxRegistry[c.ref] = r
	}
	r.refCount++
	ctxMutex.Unlock()
}

func (c *Context) deregister() {
	ctxMutex.Lock()
	defer ctxMutex.Unlock()
	r := ctxRegistry[c.ref]
	if r == nil {
		return
	}
	r.refCount--
	if r.refCount <= 0 {
		delete(ctxRegistry, c.ref)
	}
}

func getContext(ref int) *Context {
	ctxMutex.RLock()
	defer ctxMutex.RUnlock()
	r := ctxRegistry[ref]
	if r == nil {
		return nil
	}
	return r.ctx
}

//export goContext
func goContext(ref int) C.ContextPtr {
	ctx := getContext(ref)
	return ctx.ptr
}

func getValue(ctx *Context, rtn C.RtnValue) *Value {
	if rtn.value == nil {
		return nil
	}
	return &Value{rtn.value, ctx}
}

func getError(rtn C.RtnValue) error {
	if rtn.error.msg == nil {
		return nil
	}
	err := &JSError{
		Message:    C.GoString(rtn.error.msg),
		Location:   C.GoString(rtn.error.location),
		StackTrace: C.GoString(rtn.error.stack),
	}
	C.free(unsafe.Pointer(rtn.error.msg))
	C.free(unsafe.Pointer(rtn.error.location))
	C.free(unsafe.Pointer(rtn.error.stack))
	return err
}
