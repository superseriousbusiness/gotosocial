// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

#include "v8go.h"

#include <stdio.h>

#include <cstdlib>
#include <cstring>
#include <iostream>
#include <sstream>
#include <string>
#include <vector>

#if defined(__MINGW32__) || defined(__MINGW64__)
// MinGW header files do not implicitly include windows.h
struct _EXCEPTION_POINTERS;
#endif

#include "libplatform/libplatform.h"
#include "v8.h"

using namespace v8;

auto default_platform = platform::NewDefaultPlatform();
auto default_allocator = ArrayBuffer::Allocator::NewDefaultAllocator();

typedef struct {
  Isolate* iso;
  std::vector<ValuePtr> vals;
  Persistent<Context> ptr;
} m_ctx;

typedef struct {
  Isolate* iso;
  m_ctx* ctx;
  Persistent<Value, CopyablePersistentTraits<Value>> ptr;
} m_value;

typedef struct {
  Isolate* iso;
  Persistent<Template> ptr;
} m_template;

const char* CopyString(std::string str) {
  int len = str.length();
  char* mem = (char*)malloc(len + 1);
  memcpy(mem, str.data(), len);
  mem[len] = 0;
  return mem;
}

const char* CopyString(String::Utf8Value& value) {
  if (value.length() == 0) {
    return nullptr;
  }
  return CopyString(*value);
}

RtnError ExceptionError(TryCatch& try_catch, Isolate* iso, Local<Context> ctx) {
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  RtnError rtn = {nullptr, nullptr, nullptr};

  if (try_catch.HasTerminated()) {
    rtn.msg =
        CopyString("ExecutionTerminated: script execution has been terminated");
    return rtn;
  }

  String::Utf8Value exception(iso, try_catch.Exception());
  rtn.msg = CopyString(exception);

  Local<Message> msg = try_catch.Message();
  if (!msg.IsEmpty()) {
    String::Utf8Value origin(iso, msg->GetScriptOrigin().ResourceName());
    std::ostringstream sb;
    sb << *origin;
    Maybe<int> line = try_catch.Message()->GetLineNumber(ctx);
    if (line.IsJust()) {
      sb << ":" << line.ToChecked();
    }
    Maybe<int> start = try_catch.Message()->GetStartColumn(ctx);
    if (start.IsJust()) {
      sb << ":"
         << start.ToChecked() + 1;  // + 1 to match output from stack trace
    }
    rtn.location = CopyString(sb.str());
  }

  MaybeLocal<Value> mstack = try_catch.StackTrace(ctx);
  if (!mstack.IsEmpty()) {
    String::Utf8Value stack(iso, mstack.ToLocalChecked());
    rtn.stack = CopyString(stack);
  }

  return rtn;
}

ValuePtr tracked_value(m_ctx* ctx, m_value* val) {
  // (rogchap) we track values against a context so that when the context is
  // closed (either manually or GC'd by Go) we can also release all the
  // values associated with the context; previously the Go GC would not run
  // quickly enough, as it has no understanding of the C memory allocation size.
  // By doing so we hold pointers to all values that are created/returned to Go
  // until the context is released; this is a compromise.
  // Ideally we would be able to delete the value object and cancel the
  // finalizer on the Go side, but we currently don't pass the Go ptr, but
  // rather the C ptr. A potential future iteration would be to use an
  // unordered_map, where we could do O(1) lookups for the value, but then know
  // if the object has been finalized or not by being in the map or not. This
  // would require some ref id for the value rather than passing the ptr between
  // Go <--> C, which would be a significant change, as there are places where
  // we get the context from the value, but if we then need the context to get
  // the value, we would be in a circular bind.
  ValuePtr val_ptr = static_cast<ValuePtr>(val);
  ctx->vals.push_back(val_ptr);

  return val_ptr;
}

extern "C" {

/********** Isolate **********/

#define ISOLATE_SCOPE(iso_ptr)                   \
  Isolate* iso = static_cast<Isolate*>(iso_ptr); \
  Locker locker(iso);                            \
  Isolate::Scope isolate_scope(iso);             \
  HandleScope handle_scope(iso);

void Init() {
#ifdef _WIN32
  V8::InitializeExternalStartupData(".");
#endif
  V8::InitializePlatform(default_platform.get());
  V8::Initialize();
  return;
}

IsolatePtr NewIsolate() {
  Isolate::CreateParams params;
  params.array_buffer_allocator = default_allocator;
  Isolate* iso = Isolate::New(params);
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  iso->SetCaptureStackTraceForUncaughtExceptions(true);

  // Create a Context for internal use
  m_ctx* ctx = new m_ctx;
  ctx->ptr.Reset(iso, Context::New(iso));
  ctx->iso = iso;
  iso->SetData(0, ctx);

  return static_cast<IsolatePtr>(iso);
}

void IsolatePerformMicrotaskCheckpoint(IsolatePtr ptr) {
  ISOLATE_SCOPE(ptr)
  iso->PerformMicrotaskCheckpoint();
}

void IsolateDispose(IsolatePtr ptr) {
  if (ptr == nullptr) {
    return;
  }
  Isolate* iso = static_cast<Isolate*>(ptr);
  ContextFree(iso->GetData(0));

  iso->Dispose();
}

void IsolateTerminateExecution(IsolatePtr ptr) {
  Isolate* iso = static_cast<Isolate*>(ptr);
  iso->TerminateExecution();
}

IsolateHStatistics IsolationGetHeapStatistics(IsolatePtr ptr) {
  if (ptr == nullptr) {
    return IsolateHStatistics{0};
  }
  Isolate* iso = static_cast<Isolate*>(ptr);
  v8::HeapStatistics hs;
  iso->GetHeapStatistics(&hs);

  return IsolateHStatistics{hs.total_heap_size(),
                            hs.total_heap_size_executable(),
                            hs.total_physical_size(),
                            hs.total_available_size(),
                            hs.used_heap_size(),
                            hs.heap_size_limit(),
                            hs.malloced_memory(),
                            hs.external_memory(),
                            hs.peak_malloced_memory(),
                            hs.number_of_native_contexts(),
                            hs.number_of_detached_contexts()};
}

/********** Template **********/

#define LOCAL_TEMPLATE(ptr)                       \
  m_template* ot = static_cast<m_template*>(ptr); \
  Isolate* iso = ot->iso;                         \
  Locker locker(iso);                             \
  Isolate::Scope isolate_scope(iso);              \
  HandleScope handle_scope(iso);                  \
  Local<Template> tmpl = ot->ptr.Get(iso);

void TemplateFree(TemplatePtr ptr) {
  delete static_cast<m_template*>(ptr);
}

void TemplateSetValue(TemplatePtr ptr,
                      const char* name,
                      ValuePtr val_ptr,
                      int attributes) {
  LOCAL_TEMPLATE(ptr);

  Local<String> prop_name =
      String::NewFromUtf8(iso, name, NewStringType::kNormal).ToLocalChecked();
  m_value* val = static_cast<m_value*>(val_ptr);
  tmpl->Set(prop_name, val->ptr.Get(iso), (PropertyAttribute)attributes);
}

void TemplateSetTemplate(TemplatePtr ptr,
                         const char* name,
                         TemplatePtr obj_ptr,
                         int attributes) {
  LOCAL_TEMPLATE(ptr);

  Local<String> prop_name =
      String::NewFromUtf8(iso, name, NewStringType::kNormal).ToLocalChecked();
  m_template* obj = static_cast<m_template*>(obj_ptr);
  tmpl->Set(prop_name, obj->ptr.Get(iso), (PropertyAttribute)attributes);
}

/********** ObjectTemplate **********/

TemplatePtr NewObjectTemplate(IsolatePtr iso_ptr) {
  Isolate* iso = static_cast<Isolate*>(iso_ptr);
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  m_template* ot = new m_template;
  ot->iso = iso;
  ot->ptr.Reset(iso, ObjectTemplate::New(iso));
  return static_cast<TemplatePtr>(ot);
}

ValuePtr ObjectTemplateNewInstance(TemplatePtr ptr, ContextPtr ctx_ptr) {
  LOCAL_TEMPLATE(ptr);
  m_ctx* ctx = static_cast<m_ctx*>(ctx_ptr);
  Local<Context> local_ctx = ctx->ptr.Get(iso);
  Context::Scope context_scope(local_ctx);

  Local<ObjectTemplate> obj_tmpl = tmpl.As<ObjectTemplate>();
  MaybeLocal<Object> obj = obj_tmpl->NewInstance(local_ctx);

  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, obj.ToLocalChecked());
  return tracked_value(ctx, val);
}

/********** FunctionTemplate **********/

static void FunctionTemplateCallback(const FunctionCallbackInfo<Value>& info) {
  Isolate* iso_ptr = info.GetIsolate();
  ISOLATE_SCOPE(iso_ptr);

  // This callback function can be called from any Context, which we only know
  // at runtime. We extract the Context reference from the embedder data so that
  // we can use the context registry to match the Context on the Go side
  Local<Context> local_ctx = iso->GetCurrentContext();
  int ctx_ref = local_ctx->GetEmbedderData(1).As<Integer>()->Value();
  ContextPtr goContext(int ctxref);
  ContextPtr ctx_ptr = goContext(ctx_ref);
  m_ctx* ctx = static_cast<m_ctx*>(ctx_ptr);

  int callback_ref = info.Data().As<Integer>()->Value();

  int args_count = info.Length();
  ValuePtr args[args_count];
  for (int i = 0; i < args_count; i++) {
    m_value* val = new m_value;
    val->iso = iso;
    val->ctx = ctx;
    val->ptr.Reset(
        iso, Persistent<Value, CopyablePersistentTraits<Value>>(iso, info[i]));
    args[i] = tracked_value(ctx, val);
  }

  ValuePtr goFunctionCallback(int ctxref, int cbref, const ValuePtr* args,
                              int args_count);
  ValuePtr val_ptr =
      goFunctionCallback(ctx_ref, callback_ref, args, args_count);
  if (val_ptr != nullptr) {
    m_value* val = static_cast<m_value*>(val_ptr);
    info.GetReturnValue().Set(val->ptr.Get(iso));
  } else {
    info.GetReturnValue().SetUndefined();
  }
}

TemplatePtr NewFunctionTemplate(IsolatePtr iso_ptr, int callback_ref) {
  Isolate* iso = static_cast<Isolate*>(iso_ptr);
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  // (rogchap) We only need to store one value, callback_ref, into the
  // C++ callback function data, but if we needed to store more items we could
  // use an V8::Array; this would require the internal context from
  // iso->GetData(0)
  Local<Integer> cbData = Integer::New(iso, callback_ref);

  m_template* ot = new m_template;
  ot->iso = iso;
  ot->ptr.Reset(iso,
                FunctionTemplate::New(iso, FunctionTemplateCallback, cbData));
  return static_cast<TemplatePtr>(ot);
}

ValuePtr FunctionTemplateGetFunction(TemplatePtr ptr, ContextPtr ctx_ptr) {
  LOCAL_TEMPLATE(ptr);
  m_ctx* ctx = static_cast<m_ctx*>(ctx_ptr);
  Local<Context> local_ctx = ctx->ptr.Get(iso);
  Context::Scope context_scope(local_ctx);

  Local<FunctionTemplate> fn_tmpl = tmpl.As<FunctionTemplate>();
  MaybeLocal<Function> fn = fn_tmpl->GetFunction(local_ctx);

  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(iso, fn.ToLocalChecked());
  return tracked_value(ctx, val);
}

/********** Context **********/

#define LOCAL_CONTEXT(ctx_ptr)                  \
  m_ctx* ctx = static_cast<m_ctx*>(ctx_ptr);    \
  Isolate* iso = ctx->iso;                      \
  Locker locker(iso);                           \
  Isolate::Scope isolate_scope(iso);            \
  HandleScope handle_scope(iso);                \
  TryCatch try_catch(iso);                      \
  Local<Context> local_ctx = ctx->ptr.Get(iso); \
  Context::Scope context_scope(local_ctx);

ContextPtr NewContext(IsolatePtr iso_ptr,
                      TemplatePtr global_template_ptr,
                      int ref) {
  Isolate* iso = static_cast<Isolate*>(iso_ptr);
  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  Local<ObjectTemplate> global_template;
  if (global_template_ptr != nullptr) {
    m_template* ob = static_cast<m_template*>(global_template_ptr);
    global_template = ob->ptr.Get(iso).As<ObjectTemplate>();
  } else {
    global_template = ObjectTemplate::New(iso);
  }

  // For function callbacks we need a reference to the context, but because of
  // the complexities of C -> Go function pointers, we store a reference to the
  // context as a simple integer identifier; this can then be used on the Go
  // side to lookup the context in the context registry. We use slot 1 as slot 0
  // has special meaning for the Chrome debugger.
  Local<Context> local_ctx = Context::New(iso, nullptr, global_template);
  local_ctx->SetEmbedderData(1, Integer::New(iso, ref));

  m_ctx* ctx = new m_ctx;
  ctx->ptr.Reset(iso, local_ctx);
  ctx->iso = iso;
  return static_cast<ContextPtr>(ctx);
}

void ContextFree(ContextPtr ptr) {
  if (ptr == nullptr) {
    return;
  }
  m_ctx* ctx = static_cast<m_ctx*>(ptr);
  if (ctx == nullptr) {
    return;
  }
  ctx->ptr.Reset();

  for (ValuePtr val_ptr : ctx->vals) {
    ValueFree(val_ptr);
  }

  delete ctx;
}

RtnValue RunScript(ContextPtr ctx_ptr, const char* source, const char* origin) {
  LOCAL_CONTEXT(ctx_ptr);

  Local<String> src =
      String::NewFromUtf8(iso, source, NewStringType::kNormal).ToLocalChecked();
  Local<String> ogn =
      String::NewFromUtf8(iso, origin, NewStringType::kNormal).ToLocalChecked();

  RtnValue rtn = {nullptr, nullptr};

  ScriptOrigin script_origin(ogn);
  MaybeLocal<Script> script = Script::Compile(local_ctx, src, &script_origin);
  if (script.IsEmpty()) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  MaybeLocal<Value> result = script.ToLocalChecked()->Run(local_ctx);
  if (result.IsEmpty()) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, result.ToLocalChecked());

  rtn.value = tracked_value(ctx, val);
  return rtn;
}

RtnValue JSONParse(ContextPtr ctx_ptr, const char* str) {
  LOCAL_CONTEXT(ctx_ptr);
  RtnValue rtn = {nullptr, nullptr};

  MaybeLocal<Value> result = JSON::Parse(
      local_ctx,
      String::NewFromUtf8(iso, str, NewStringType::kNormal).ToLocalChecked());
  if (result.IsEmpty()) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, result.ToLocalChecked());

  rtn.value = tracked_value(ctx, val);
  return rtn;
}

const char* JSONStringify(ContextPtr ctx_ptr, ValuePtr val_ptr) {
  Isolate* iso;
  Local<Context> local_ctx;

  m_value* val = static_cast<m_value*>(val_ptr);
  m_ctx* ctx = static_cast<m_ctx*>(ctx_ptr);

  if (ctx != nullptr) {
    iso = ctx->iso;
  } else {
    iso = val->iso;
  }

  Locker locker(iso);
  Isolate::Scope isolate_scope(iso);
  HandleScope handle_scope(iso);

  if (ctx != nullptr) {
    local_ctx = ctx->ptr.Get(iso);
  } else {
    if (val->ctx != nullptr) {
      local_ctx = val->ctx->ptr.Get(iso);
    } else {
      m_ctx* ctx = static_cast<m_ctx*>(iso->GetData(0));
      local_ctx = ctx->ptr.Get(iso);
    }
  }

  Context::Scope context_scope(local_ctx);

  MaybeLocal<String> str = JSON::Stringify(local_ctx, val->ptr.Get(iso));
  if (str.IsEmpty()) {
    return nullptr;
  }
  String::Utf8Value json(iso, str.ToLocalChecked());
  return CopyString(json);
}

ValuePtr ContextGlobal(ContextPtr ctx_ptr) {
  LOCAL_CONTEXT(ctx_ptr);
  m_value* val = new m_value;

  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, local_ctx->Global());

  return tracked_value(ctx, val);
}

/********** Value **********/

#define LOCAL_VALUE(ptr)                        \
  m_value* val = static_cast<m_value*>(ptr);    \
  Isolate* iso = val->iso;                      \
  Locker locker(iso);                           \
  Isolate::Scope isolate_scope(iso);            \
  HandleScope handle_scope(iso);                \
  TryCatch try_catch(iso);                      \
  m_ctx* ctx = val->ctx;                        \
  Local<Context> local_ctx;                     \
  if (ctx != nullptr) {                         \
    local_ctx = ctx->ptr.Get(iso);              \
  } else {                                      \
    ctx = static_cast<m_ctx*>(iso->GetData(0)); \
    local_ctx = ctx->ptr.Get(iso);              \
  }                                             \
  Context::Scope context_scope(local_ctx);      \
  Local<Value> value = val->ptr.Get(iso);

#define ISOLATE_SCOPE_INTERNAL_CONTEXT(iso_ptr) \
  ISOLATE_SCOPE(iso_ptr);                       \
  m_ctx* ctx = static_cast<m_ctx*>(iso->GetData(0));

ValuePtr NewValueInteger(IsolatePtr iso_ptr, int32_t v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso_ptr);
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, Integer::New(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueIntegerFromUnsigned(IsolatePtr iso_ptr, uint32_t v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso_ptr);
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, Integer::NewFromUnsigned(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueString(IsolatePtr iso_ptr, const char* v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso_ptr);
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, String::NewFromUtf8(iso, v).ToLocalChecked());
  return tracked_value(ctx, val);
}

ValuePtr NewValueBoolean(IsolatePtr iso_ptr, int v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso_ptr);
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, Boolean::New(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueNumber(IsolatePtr iso_ptr, double v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso_ptr);
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, Number::New(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueBigInt(IsolatePtr iso_ptr, int64_t v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso_ptr);
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, BigInt::New(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueBigIntFromUnsigned(IsolatePtr iso_ptr, uint64_t v) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso_ptr);
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, BigInt::NewFromUnsigned(iso, v));
  return tracked_value(ctx, val);
}

ValuePtr NewValueBigIntFromWords(IsolatePtr iso_ptr,
                                 int sign_bit,
                                 int word_count,
                                 const uint64_t* words) {
  ISOLATE_SCOPE_INTERNAL_CONTEXT(iso_ptr);

  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  MaybeLocal<BigInt> bigint =
      BigInt::NewFromWords(ctx->ptr.Get(iso), sign_bit, word_count, words);
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, bigint.ToLocalChecked());
  return tracked_value(ctx, val);
}

void ValueFree(ValuePtr ptr) {
  if (ptr == nullptr) {
    return;
  }
  m_value* val = static_cast<m_value*>(ptr);
  val->ptr.Reset();
  delete val;
}

const uint32_t* ValueToArrayIndex(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  MaybeLocal<Uint32> array_index = value->ToArrayIndex(local_ctx);
  if (array_index.IsEmpty()) {
    return nullptr;
  }

  uint32_t* idx = new uint32_t;
  *idx = array_index.ToLocalChecked()->Value();
  return idx;
}

int ValueToBoolean(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->BooleanValue(iso);
}

int32_t ValueToInt32(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->Int32Value(local_ctx).ToChecked();
}

int64_t ValueToInteger(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IntegerValue(local_ctx).ToChecked();
}

double ValueToNumber(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->NumberValue(local_ctx).ToChecked();
}

const char* ValueToDetailString(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  String::Utf8Value ds(iso, value->ToDetailString(local_ctx).ToLocalChecked());
  return CopyString(ds);
}

const char* ValueToString(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  String::Utf8Value utf8(iso, value);
  return CopyString(utf8);
}

uint32_t ValueToUint32(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->Uint32Value(local_ctx).ToChecked();
}

ValueBigInt ValueToBigInt(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  MaybeLocal<BigInt> bint = value->ToBigInt(local_ctx);
  if (bint.IsEmpty()) {
    return {nullptr, 0};
  }

  int word_count = bint.ToLocalChecked()->WordCount();
  int sign_bit = 0;
  uint64_t* words = new uint64_t[word_count];
  bint.ToLocalChecked()->ToWordsArray(&sign_bit, &word_count, words);
  ValueBigInt rtn = {words, word_count, sign_bit};
  return rtn;
}

ValuePtr ValueToObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  m_value* new_val = new m_value;
  new_val->iso = iso;
  new_val->ctx = ctx;
  new_val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, value->ToObject(local_ctx).ToLocalChecked());
  return tracked_value(ctx, new_val);
}

int ValueIsUndefined(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUndefined();
}

int ValueIsNull(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNull();
}

int ValueIsNullOrUndefined(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNullOrUndefined();
}

int ValueIsTrue(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsTrue();
}

int ValueIsFalse(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsFalse();
}

int ValueIsName(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsName();
}

int ValueIsString(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsString();
}

int ValueIsSymbol(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSymbol();
}

int ValueIsFunction(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsFunction();
}

int ValueIsObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsObject();
}

int ValueIsBigInt(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBigInt();
}

int ValueIsBoolean(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBoolean();
}

int ValueIsNumber(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNumber();
}

int ValueIsExternal(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsExternal();
}

int ValueIsInt32(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsInt32();
}

int ValueIsUint32(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint32();
}

int ValueIsDate(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsDate();
}

int ValueIsArgumentsObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsArgumentsObject();
}

int ValueIsBigIntObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBigIntObject();
}

int ValueIsNumberObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNumberObject();
}

int ValueIsStringObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsStringObject();
}

int ValueIsSymbolObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSymbolObject();
}

int ValueIsNativeError(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsNativeError();
}

int ValueIsRegExp(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsRegExp();
}

int ValueIsAsyncFunction(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsAsyncFunction();
}

int ValueIsGeneratorFunction(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsGeneratorFunction();
}

int ValueIsGeneratorObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsGeneratorObject();
}

int ValueIsPromise(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsPromise();
}

int ValueIsMap(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsMap();
}

int ValueIsSet(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSet();
}

int ValueIsMapIterator(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsMapIterator();
}

int ValueIsSetIterator(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSetIterator();
}

int ValueIsWeakMap(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsWeakMap();
}

int ValueIsWeakSet(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsWeakSet();
}

int ValueIsArray(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsArray();
}

int ValueIsArrayBuffer(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsArrayBuffer();
}

int ValueIsArrayBufferView(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsArrayBufferView();
}

int ValueIsTypedArray(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsTypedArray();
}

int ValueIsUint8Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint8Array();
}

int ValueIsUint8ClampedArray(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint8ClampedArray();
}

int ValueIsInt8Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsInt8Array();
}

int ValueIsUint16Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint16Array();
}

int ValueIsInt16Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsInt16Array();
}

int ValueIsUint32Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsUint32Array();
}

int ValueIsInt32Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsInt32Array();
}

int ValueIsFloat32Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsFloat32Array();
}

int ValueIsFloat64Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsFloat64Array();
}

int ValueIsBigInt64Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBigInt64Array();
}

int ValueIsBigUint64Array(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsBigUint64Array();
}

int ValueIsDataView(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsDataView();
}

int ValueIsSharedArrayBuffer(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsSharedArrayBuffer();
}

int ValueIsProxy(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsProxy();
}

int ValueIsWasmModuleObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsWasmModuleObject();
}

int ValueIsModuleNamespaceObject(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  return value->IsModuleNamespaceObject();
}

/********** Object **********/

#define LOCAL_OBJECT(ptr) \
  LOCAL_VALUE(ptr)        \
  Local<Object> obj = value.As<Object>()

void ObjectSet(ValuePtr ptr, const char* key, ValuePtr val_ptr) {
  LOCAL_OBJECT(ptr);
  Local<String> key_val =
      String::NewFromUtf8(iso, key, NewStringType::kNormal).ToLocalChecked();
  m_value* prop_val = static_cast<m_value*>(val_ptr);
  obj->Set(local_ctx, key_val, prop_val->ptr.Get(iso)).Check();
}

void ObjectSetIdx(ValuePtr ptr, uint32_t idx, ValuePtr val_ptr) {
  LOCAL_OBJECT(ptr);
  m_value* prop_val = static_cast<m_value*>(val_ptr);
  obj->Set(local_ctx, idx, prop_val->ptr.Get(iso)).Check();
}

RtnValue ObjectGet(ValuePtr ptr, const char* key) {
  LOCAL_OBJECT(ptr);
  RtnValue rtn = {nullptr, nullptr};

  Local<String> key_val =
      String::NewFromUtf8(iso, key, NewStringType::kNormal).ToLocalChecked();
  MaybeLocal<Value> result = obj->Get(local_ctx, key_val);
  if (result.IsEmpty()) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* new_val = new m_value;
  new_val->iso = iso;
  new_val->ctx = ctx;
  new_val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, result.ToLocalChecked());

  rtn.value = tracked_value(ctx, new_val);
  return rtn;
}

RtnValue ObjectGetIdx(ValuePtr ptr, uint32_t idx) {
  LOCAL_OBJECT(ptr);
  RtnValue rtn = {nullptr, nullptr};

  MaybeLocal<Value> result = obj->Get(local_ctx, idx);
  if (result.IsEmpty()) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* new_val = new m_value;
  new_val->iso = iso;
  new_val->ctx = ctx;
  new_val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, result.ToLocalChecked());

  rtn.value = tracked_value(ctx, new_val);
  return rtn;
}

int ObjectHas(ValuePtr ptr, const char* key) {
  LOCAL_OBJECT(ptr);
  Local<String> key_val =
      String::NewFromUtf8(iso, key, NewStringType::kNormal).ToLocalChecked();
  return obj->Has(local_ctx, key_val).ToChecked();
}

int ObjectHasIdx(ValuePtr ptr, uint32_t idx) {
  LOCAL_OBJECT(ptr);
  return obj->Has(local_ctx, idx).ToChecked();
}

int ObjectDelete(ValuePtr ptr, const char* key) {
  LOCAL_OBJECT(ptr);
  Local<String> key_val =
      String::NewFromUtf8(iso, key, NewStringType::kNormal).ToLocalChecked();
  return obj->Delete(local_ctx, key_val).ToChecked();
}

int ObjectDeleteIdx(ValuePtr ptr, uint32_t idx) {
  LOCAL_OBJECT(ptr);
  return obj->Delete(local_ctx, idx).ToChecked();
}

/********** Promise **********/

ValuePtr NewPromiseResolver(ContextPtr ctx_ptr) {
  LOCAL_CONTEXT(ctx_ptr);
  MaybeLocal<Promise::Resolver> resolver = Promise::Resolver::New(local_ctx);
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = ctx;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, resolver.ToLocalChecked());
  return tracked_value(ctx, val);
}

ValuePtr PromiseResolverGetPromise(ValuePtr ptr) {
  LOCAL_VALUE(ptr);
  Local<Promise::Resolver> resolver = value.As<Promise::Resolver>();
  Local<Promise> promise = resolver->GetPromise();
  m_value* promise_val = new m_value;
  promise_val->iso = iso;
  promise_val->ctx = ctx;
  promise_val->ptr =
      Persistent<Value, CopyablePersistentTraits<Value>>(iso, promise);
  return tracked_value(ctx, promise_val);
}

int PromiseResolverResolve(ValuePtr ptr, ValuePtr val_ptr) {
  LOCAL_VALUE(ptr);
  Local<Promise::Resolver> resolver = value.As<Promise::Resolver>();
  m_value* resolve_val = static_cast<m_value*>(val_ptr);
  return resolver->Resolve(local_ctx, resolve_val->ptr.Get(iso)).ToChecked();
}

int PromiseResolverReject(ValuePtr ptr, ValuePtr val_ptr) {
  LOCAL_VALUE(ptr);
  Local<Promise::Resolver> resolver = value.As<Promise::Resolver>();
  m_value* reject_val = static_cast<m_value*>(val_ptr);
  return resolver->Reject(local_ctx, reject_val->ptr.Get(iso)).ToChecked();
}

int PromiseState(ValuePtr ptr) {
  LOCAL_VALUE(ptr)
  Local<Promise> promise = value.As<Promise>();
  return promise->State();
}

ValuePtr PromiseThen(ValuePtr ptr, int callback_ref) {
  LOCAL_VALUE(ptr)
  Local<Promise> promise = value.As<Promise>();
  Local<Integer> cbData = Integer::New(iso, callback_ref);
  Local<Function> func = Function::New(local_ctx, FunctionTemplateCallback, cbData)
    .ToLocalChecked();
  Local<Promise> result = promise->Then(local_ctx, func).ToLocalChecked();
  m_value* promise_val = new m_value;
  promise_val->iso = iso;
  promise_val->ctx = ctx;
  promise_val->ptr =
      Persistent<Value, CopyablePersistentTraits<Value>>(iso, promise);
  return tracked_value(ctx, promise_val);
}

ValuePtr PromiseThen2(ValuePtr ptr, int on_fulfilled_ref, int on_rejected_ref) {
  LOCAL_VALUE(ptr)
  Local<Promise> promise = value.As<Promise>();
  Local<Integer> onFulfilledData = Integer::New(iso, on_fulfilled_ref);
  Local<Function> onFulfilledFunc = Function::New(local_ctx, FunctionTemplateCallback, onFulfilledData)
    .ToLocalChecked();
  Local<Integer> onRejectedData = Integer::New(iso, on_rejected_ref);
  Local<Function> onRejectedFunc = Function::New(local_ctx, FunctionTemplateCallback, onRejectedData)
    .ToLocalChecked();
  Local<Promise> result = promise->Then(local_ctx, onFulfilledFunc, onRejectedFunc).ToLocalChecked();
  m_value* promise_val = new m_value;
  promise_val->iso = iso;
  promise_val->ctx = ctx;
  promise_val->ptr =
      Persistent<Value, CopyablePersistentTraits<Value>>(iso, promise);
  return tracked_value(ctx, promise_val);
}

ValuePtr PromiseCatch(ValuePtr ptr, int callback_ref) {
  LOCAL_VALUE(ptr)
  Local<Promise> promise = value.As<Promise>();
  Local<Integer> cbData = Integer::New(iso, callback_ref);
  Local<Function> func = Function::New(local_ctx, FunctionTemplateCallback, cbData)
    .ToLocalChecked();
  Local<Promise> result = promise->Catch(local_ctx, func).ToLocalChecked();
  m_value* promise_val = new m_value;
  promise_val->iso = iso;
  promise_val->ctx = ctx;
  promise_val->ptr =
      Persistent<Value, CopyablePersistentTraits<Value>>(iso, promise);
  return tracked_value(ctx, promise_val);
}

ValuePtr PromiseResult(ValuePtr ptr) {
  LOCAL_VALUE(ptr)
  Local<Promise> promise = value.As<Promise>();
  Local<Value> result = promise->Result();
  m_value* result_val = new m_value;
  result_val->iso = iso;
  result_val->ctx = ctx;
  result_val->ptr =
      Persistent<Value, CopyablePersistentTraits<Value>>(iso, result);
  return tracked_value(ctx, result_val);
}

/********** Function **********/

RtnValue FunctionCall(ValuePtr ptr, int argc, ValuePtr args[]) {
  LOCAL_VALUE(ptr)
  RtnValue rtn = {nullptr, nullptr};
  Local<Function> fn = Local<Function>::Cast(value);
  Local<Value> argv[argc];
  for (int i = 0; i < argc; i++) {
    m_value* arg = static_cast<m_value*>(args[i]);
    argv[i] = arg->ptr.Get(iso);
  }
  Local<Value> recv = Undefined(iso);
  MaybeLocal<Value> result = fn->Call(local_ctx, recv, argc, argv);
  if (result.IsEmpty()) {
    rtn.error = ExceptionError(try_catch, iso, local_ctx);
    return rtn;
  }
  m_value* rtnval = new m_value;
  rtnval->iso = iso;
  rtnval->ctx = ctx;
  rtnval->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(iso, result.ToLocalChecked());
  rtn.value = tracked_value(ctx, rtnval);
  return rtn;
}

/******** Exceptions *********/

ValuePtr ExceptionError(IsolatePtr iso_ptr, const char* message) {
  ISOLATE_SCOPE(iso_ptr);
  Local<String> msg = String::NewFromUtf8(iso, message).ToLocalChecked();
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = nullptr;
  // TODO(rogchap): This currently causes a segfault, and I'm not sure why!
  // Even a simple error with an empty string causes the error:
  // Exception::Error(String::Empty(iso))
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, Exception::Error(msg));
  return static_cast<ValuePtr>(val);
}

ValuePtr ExceptionRangeError(IsolatePtr iso_ptr, const char* message) {
  ISOLATE_SCOPE(iso_ptr);
  Local<String> msg = String::NewFromUtf8(iso, message, NewStringType::kNormal)
                          .ToLocalChecked();
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = nullptr;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, Exception::RangeError(msg));
  return static_cast<ValuePtr>(val);
}

ValuePtr ExceptionReferenceError(IsolatePtr iso_ptr, const char* message) {
  ISOLATE_SCOPE(iso_ptr);
  Local<String> msg = String::NewFromUtf8(iso, message, NewStringType::kNormal)
                          .ToLocalChecked();
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = nullptr;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, Exception::ReferenceError(msg));
  return static_cast<ValuePtr>(val);
}

ValuePtr ExceptionSyntaxError(IsolatePtr iso_ptr, const char* message) {
  ISOLATE_SCOPE(iso_ptr);
  Local<String> msg = String::NewFromUtf8(iso, message, NewStringType::kNormal)
                          .ToLocalChecked();
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = nullptr;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, Exception::SyntaxError(msg));
  return static_cast<ValuePtr>(val);
}

ValuePtr ExceptionTypeError(IsolatePtr iso_ptr, const char* message) {
  ISOLATE_SCOPE(iso_ptr);
  Local<String> msg = String::NewFromUtf8(iso, message, NewStringType::kNormal)
                          .ToLocalChecked();
  m_value* val = new m_value;
  val->iso = iso;
  val->ctx = nullptr;
  val->ptr = Persistent<Value, CopyablePersistentTraits<Value>>(
      iso, Exception::TypeError(msg));
  return static_cast<ValuePtr>(val);
}

/********** v8::V8 **********/

const char* Version() {
  return V8::GetVersion();
}

void SetFlags(const char* flags) {
  V8::SetFlagsFromString(flags);
}
}
