// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

#ifndef V8GO_H
#define V8GO_H
#ifdef __cplusplus

extern "C" {
#endif

#include <stddef.h>
#include <stdint.h>

typedef void* IsolatePtr;
typedef void* ContextPtr;
typedef void* ValuePtr;
typedef void* TemplatePtr;

typedef struct {
  const char* msg;
  const char* location;
  const char* stack;
} RtnError;

typedef struct {
  ValuePtr value;
  RtnError error;
} RtnValue;

typedef struct {
  size_t total_heap_size;
  size_t total_heap_size_executable;
  size_t total_physical_size;
  size_t total_available_size;
  size_t used_heap_size;
  size_t heap_size_limit;
  size_t malloced_memory;
  size_t external_memory;
  size_t peak_malloced_memory;
  size_t number_of_native_contexts;
  size_t number_of_detached_contexts;
} IsolateHStatistics;

typedef struct {
  const uint64_t* word_array;
  int word_count;
  int sign_bit;
} ValueBigInt;

extern void Init();
extern IsolatePtr NewIsolate();
extern void IsolatePerformMicrotaskCheckpoint(IsolatePtr ptr);
extern void IsolateDispose(IsolatePtr ptr);
extern void IsolateTerminateExecution(IsolatePtr ptr);
extern IsolateHStatistics IsolationGetHeapStatistics(IsolatePtr ptr);

extern ContextPtr NewContext(IsolatePtr iso_ptr,
                             TemplatePtr global_template_ptr,
                             int ref);
extern void ContextFree(ContextPtr ptr);
extern RtnValue RunScript(ContextPtr ctx_ptr,
                          const char* source,
                          const char* origin);
extern RtnValue JSONParse(ContextPtr ctx_ptr, const char* str);
const char* JSONStringify(ContextPtr ctx_ptr, ValuePtr val_ptr);
extern ValuePtr ContextGlobal(ContextPtr ctx_ptr);

extern void TemplateFree(TemplatePtr ptr);
extern void TemplateSetValue(TemplatePtr ptr,
                             const char* name,
                             ValuePtr val_ptr,
                             int attributes);
extern void TemplateSetTemplate(TemplatePtr ptr,
                                const char* name,
                                TemplatePtr obj_ptr,
                                int attributes);

extern TemplatePtr NewObjectTemplate(IsolatePtr iso_ptr);
extern ValuePtr ObjectTemplateNewInstance(TemplatePtr ptr, ContextPtr ctx_ptr);

extern TemplatePtr NewFunctionTemplate(IsolatePtr iso_ptr, int callback_ref);
extern ValuePtr FunctionTemplateGetFunction(TemplatePtr ptr, ContextPtr ctx_ptr);

extern ValuePtr NewValueInteger(IsolatePtr iso_ptr, int32_t v);
extern ValuePtr NewValueIntegerFromUnsigned(IsolatePtr iso_ptr, uint32_t v);
extern ValuePtr NewValueString(IsolatePtr iso_ptr, const char* v);
extern ValuePtr NewValueBoolean(IsolatePtr iso_ptr, int v);
extern ValuePtr NewValueNumber(IsolatePtr iso_ptr, double v);
extern ValuePtr NewValueBigInt(IsolatePtr iso_ptr, int64_t v);
extern ValuePtr NewValueBigIntFromUnsigned(IsolatePtr iso_ptr, uint64_t v);
extern ValuePtr NewValueBigIntFromWords(IsolatePtr iso_ptr,
                                        int sign_bit,
                                        int word_count,
                                        const uint64_t* words);
extern void ValueFree(ValuePtr ptr);
const char* ValueToString(ValuePtr ptr);
const uint32_t* ValueToArrayIndex(ValuePtr ptr);
int ValueToBoolean(ValuePtr ptr);
int32_t ValueToInt32(ValuePtr ptr);
int64_t ValueToInteger(ValuePtr ptr);
double ValueToNumber(ValuePtr ptr);
const char* ValueToDetailString(ValuePtr ptr);
uint32_t ValueToUint32(ValuePtr ptr);
extern ValueBigInt ValueToBigInt(ValuePtr ptr);
extern ValuePtr ValueToObject(ValuePtr ptr);
int ValueIsUndefined(ValuePtr ptr);
int ValueIsNull(ValuePtr ptr);
int ValueIsNullOrUndefined(ValuePtr ptr);
int ValueIsTrue(ValuePtr ptr);
int ValueIsFalse(ValuePtr ptr);
int ValueIsName(ValuePtr ptr);
int ValueIsString(ValuePtr ptr);
int ValueIsSymbol(ValuePtr ptr);
int ValueIsFunction(ValuePtr ptr);
int ValueIsObject(ValuePtr ptr);
int ValueIsBigInt(ValuePtr ptr);
int ValueIsBoolean(ValuePtr ptr);
int ValueIsNumber(ValuePtr ptr);
int ValueIsExternal(ValuePtr ptr);
int ValueIsInt32(ValuePtr ptr);
int ValueIsUint32(ValuePtr ptr);
int ValueIsDate(ValuePtr ptr);
int ValueIsArgumentsObject(ValuePtr ptr);
int ValueIsBigIntObject(ValuePtr ptr);
int ValueIsNumberObject(ValuePtr ptr);
int ValueIsStringObject(ValuePtr ptr);
int ValueIsSymbolObject(ValuePtr ptr);
int ValueIsNativeError(ValuePtr ptr);
int ValueIsRegExp(ValuePtr ptr);
int ValueIsAsyncFunction(ValuePtr ptr);
int ValueIsGeneratorFunction(ValuePtr ptr);
int ValueIsGeneratorObject(ValuePtr ptr);
int ValueIsPromise(ValuePtr ptr);
int ValueIsMap(ValuePtr ptr);
int ValueIsSet(ValuePtr ptr);
int ValueIsMapIterator(ValuePtr ptr);
int ValueIsSetIterator(ValuePtr ptr);
int ValueIsWeakMap(ValuePtr ptr);
int ValueIsWeakSet(ValuePtr ptr);
int ValueIsArray(ValuePtr ptr);
int ValueIsArrayBuffer(ValuePtr ptr);
int ValueIsArrayBufferView(ValuePtr ptr);
int ValueIsTypedArray(ValuePtr ptr);
int ValueIsUint8Array(ValuePtr ptr);
int ValueIsUint8ClampedArray(ValuePtr ptr);
int ValueIsInt8Array(ValuePtr ptr);
int ValueIsUint16Array(ValuePtr ptr);
int ValueIsInt16Array(ValuePtr ptr);
int ValueIsUint32Array(ValuePtr ptr);
int ValueIsInt32Array(ValuePtr ptr);
int ValueIsFloat32Array(ValuePtr ptr);
int ValueIsFloat64Array(ValuePtr ptr);
int ValueIsBigInt64Array(ValuePtr ptr);
int ValueIsBigUint64Array(ValuePtr ptr);
int ValueIsDataView(ValuePtr ptr);
int ValueIsSharedArrayBuffer(ValuePtr ptr);
int ValueIsProxy(ValuePtr ptr);
int ValueIsWasmModuleObject(ValuePtr ptr);
int ValueIsModuleNamespaceObject(ValuePtr ptr);

extern void ObjectSet(ValuePtr ptr, const char* key, ValuePtr val_ptr);
extern void ObjectSetIdx(ValuePtr ptr, uint32_t idx, ValuePtr val_ptr);
extern RtnValue ObjectGet(ValuePtr ptr, const char* key);
extern RtnValue ObjectGetIdx(ValuePtr ptr, uint32_t idx);
int ObjectHas(ValuePtr ptr, const char* key);
int ObjectHasIdx(ValuePtr ptr, uint32_t idx);
int ObjectDelete(ValuePtr ptr, const char* key);
int ObjectDeleteIdx(ValuePtr ptr, uint32_t idx);

extern ValuePtr NewPromiseResolver(ContextPtr ctx_ptr);
extern ValuePtr PromiseResolverGetPromise(ValuePtr ptr);
int PromiseResolverResolve(ValuePtr ptr, ValuePtr val_ptr);
int PromiseResolverReject(ValuePtr ptr, ValuePtr val_ptr);
int PromiseState(ValuePtr ptr);
ValuePtr PromiseThen(ValuePtr ptr, int callback_ref);
ValuePtr PromiseThen2(ValuePtr ptr, int on_fulfilled_ref, int on_rejected_ref);
ValuePtr PromiseCatch(ValuePtr ptr, int callback_ref);
extern ValuePtr PromiseResult(ValuePtr ptr);

extern RtnValue FunctionCall(ValuePtr ptr, int argc, ValuePtr argv[]);

extern ValuePtr ExceptionError(IsolatePtr iso_ptr, const char* message);
extern ValuePtr ExceptionRangeError(IsolatePtr iso_ptr, const char* message);
extern ValuePtr ExceptionReferenceError(IsolatePtr iso_ptr,
                                        const char* message);
extern ValuePtr ExceptionSyntaxError(IsolatePtr iso_ptr, const char* message);
extern ValuePtr ExceptionTypeError(IsolatePtr iso_ptr, const char* message);

const char* Version();
extern void SetFlags(const char* flags);

#ifdef __cplusplus
}  // extern "C"
#endif
#endif  // V8GO_H
