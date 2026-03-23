# Go Concepts for Kubernetes Operators

**For**: Beginners in Go working on Kubebuilder/controller projects  
**Purpose**: A standalone reference for Go concepts you will encounter in this codebase.

---

## Overview

This document explains Go language concepts used in the ShipperCD project — packages, imports, structs, pointers, error handling, and more. It is not a full Go tutorial; it focuses on ideas that appear when building Kubernetes controllers.

---

## 1. Packages

A **package** is a collection of Go files in the same directory. All files in a directory must declare the same package name.

- **`package main`** — Used for executable programs. The entry point is `func main()`.
- **Other packages** — e.g. `package v1alpha1` for API types. They provide reusable code to other packages.

**Exported vs. unexported**: Names that start with a capital letter are exported (visible to other packages). Names starting with lowercase are private to the package.

---

## 2. Imports

Imports bring code from other packages into your file.

```go
import (
    "os"                              // Standard library
    "sigs.k8s.io/controller-runtime"  // Third-party
)
```

The path in quotes is the package's import path (usually its module path + directory).

### Import for Side Effects Only

Some packages do work when loaded — their `init()` runs. You may never call anything from the package. To import without using names, use the blank identifier:

```go
_ "k8s.io/client-go/plugin/pkg/client/auth"
```

Without `_`, the compiler would report `"imported and not used"`. The `_` tells the compiler you are importing on purpose for side effects.

---

## 3. Variables

**Declaration**:
- `var x int` — declare with type
- `x := 5` — short declaration (type inferred)
- `var (a = 1; b = 2)` — multiple at once

**Package-level variables** are initialized when the package loads and shared by all code in that package.

---

## 4. Functions

```go
func name(arg1 Type1, arg2 Type2) ReturnType {
    return value
}
```

**Multiple return values** — Go functions often return `(value, error)`:

```go
result, err := someFunction()
if err != nil {
    // handle error
}
```

**No exceptions** — Errors are explicit return values.

---

## 5. Error Handling

Go does not have try/catch. Functions that can fail return `(result, error)`. You check `err` and decide what to do.

```go
mgr, err := ctrl.NewManager(...)
if err != nil {
    return err
}
```

Conventions:
- `nil` means no error
- Callers should always check `err` before using the result

---

## 6. `init()` Function

A package can define `func init()`. It runs automatically when the package is loaded, before `main()`. Used for one-time setup (e.g. registering types in a scheme).

---

## 7. Structs

A **struct** groups related fields:

```go
type EnvironmentSpec struct {
    Name      string
    Databases []string
}
```

Structs are the main way to define data shapes. In CRDs, `Spec` and `Status` are structs.

---

## 8. Struct Tags

Fields can have **tags** — metadata in backticks for serialization and tooling:

```go
Name string `json:"name"`
```

- `json:"name"` — used when marshalling/unmarshalling JSON (and YAML). **No space** after the colon: `json:"name"` ✓, `json: "name"` ✗.
- `omitempty` — omit the field when it is the zero value (empty string, nil, 0).
- `omitzero` — (Go 1.24+) similar to omitempty; omits zero values. Useful for optional fields without pointers.

**Kubebuilder markers** (for CRDs) go in **comments above** fields, not in struct tags: `// +kubebuilder:validation:Required`. Do not use `// +required` — it is not recognized by controller-gen.

---

## 9. Embedding

Go supports **embedding** — one struct can embed another to reuse its fields:

```go
type Environment struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`   // or omitzero (Go 1.24+)
    Spec   EnvironmentSpec   `json:"spec"`
    Status EnvironmentStatus `json:"status,omitempty"`
}
```

`Environment` gets all fields from `TypeMeta` and `ObjectMeta` as if they were defined on it.

---

## 10. Pointers

A **pointer** holds the **memory address** of a value, not the value itself. `*T` means "pointer to T".

### Syntax

| Notation | Meaning |
|----------|---------|
| `*T` | Type "pointer to T" (e.g. `*int` = pointer to int) |
| `&x` | Address of variable `x` |
| `*p` | Value pointed to by `p` (dereferencing) |
| `nil` | Zero value for pointers — "points to nothing" |

### Simple Example

```go
x := 42
p := &x        // p holds the address of x
*p = 100       // modify x through p
// x is now 100
```

**Without pointer** — the function receives a copy, so the original stays unchanged:

```go
func double(n int) {
    n = n * 2   // only the local copy changes
}
x := 5
double(x)      // x is still 5
```

**With pointer** — the function can modify the original:

```go
func double(n *int) {
    *n = *n * 2   // modifies the value at that address
}
x := 5
double(&x)     // x is now 10
```

### Why Use Pointers?

1. **Avoid copying large structs** — Passing `&obj` passes only an address, not the whole struct. Important for large types like `KafkaTopic` or `Deployment`.
2. **Allow mutation** — If a function needs to modify the original value, it must receive a pointer. Example: `meta.SetStatusCondition(&topic.Status.Conditions, cond)` modifies `topic.Status.Conditions` in place.
3. **Represent "optional"** — `nil` means "absent" or "not set". Useful for distinguishing "not provided" from "zero value" (e.g. `*int` can be nil vs 0).

### In This Codebase

- **Pointer receiver**: `func (r *KafkaTopicReconciler) Reconcile(...)` — `r` is a pointer so the method can modify the reconciler and avoids copying it.
- **Fill-in pattern**: `r.Get(ctx, req.NamespacedName, &kafkaTopic)` — `Get` fills the object you pass; you give `&kafkaTopic` so it can write into it.
- **Mutation**: `meta.SetStatusCondition(&kafkaTopic.Status.Conditions, ...)` — The function expects a pointer to the slice so it can update it in place.

**Mnemonic**: `&` = "take address" (for passing to a function). `*` = either the type (pointer to T) or "the value pointed to" when reading/writing.

---

## 11. Methods and Receivers

A **method** is a function attached to a type. It has access to the type's data through a **receiver**.

### Method vs Function

```go
// Function — standalone
func Reconcile(r *Reconciler, ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // called as: Reconcile(myReconciler, ctx, req)
}

// Method — attached to a type
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // called as: myReconciler.Reconcile(ctx, req)
}
```

With a method, the receiver `r` is implicit: it's the value you call the method on. `myReconciler.Reconcile(ctx, req)` passes `myReconciler` as `r` automatically.

### Receiver Syntax

The receiver appears in parentheses before the method name:

```go
func (receiverName ReceiverType) MethodName(args) ReturnType {
    // receiverName is available inside the method (like "self" or "this")
}
```

`ReceiverType` can be a **value** (`T`) or a **pointer** (`*T`).

### Value Receiver vs Pointer Receiver

| Receiver     | Syntax    | What it means                                                  |
|--------------|-----------|----------------------------------------------------------------|
| Value        | `(r Reconciler)` | Go copies the struct; changes to `r` don't affect the original |
| Pointer      | `(r *Reconciler)` | Go passes the address; changes to `r` affect the original      |

**Value receiver** — use for small, immutable types:

```go
type NamespacedName struct { Namespace, Name string }

func (n NamespacedName) String() string {
    return n.Namespace + "/" + n.Name
}
```

`NamespacedName` is small; copying is cheap. The method doesn't need to mutate it.

**Pointer receiver** — use when you need to mutate or when the type is large:

```go
func (r *KafkaTopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // r is the reconciler; can access r.Client, r.Scheme, etc.
    err := r.Get(ctx, req.NamespacedName, &kafkaTopic)  // uses r.Client
    return ctrl.Result{}, nil
}
```

The reconciler has `Client` and `Scheme`; copying it would be unnecessary and methods may need to store state. Pointer receiver is appropriate.

### Consistency Rule

If **any** method on a type uses a pointer receiver, **all** methods on that type should use pointer receivers. Go allows mixing, but the convention is to be consistent to avoid confusion about when the original is modified.

### In This Codebase

- `func (r *KafkaTopicReconciler) Reconcile(...)` — Pointer receiver: `r` holds `Client` and `Scheme`; no need to copy, and the method uses `r.Get`, `r.Status()`.
- `ctrl.NewControllerManagedBy(mgr).For(&kafkav1alpha1.KafkaTopic{})` — `For` and similar builders are methods that return the same or a new builder, allowing chaining like `.For(...).Named("...").Complete()`.

---

## 12. Context

`context.Context` is a standard way to carry deadlines, cancellation signals, and request-scoped values. It is often the first parameter of long-running or cancelable operations.

```go
func Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
```

When the context is cancelled (e.g. shutting down), code should stop work and return. Controller-runtime passes a context into `Reconcile`.

---

## 13. Interfaces

An **interface** is a contract: it lists method signatures (name, parameters, return types). Any type that has those methods **satisfies** the interface. In Go, this is **implicit** — you never write "implements"; the compiler checks it for you.

### Go vs Java (for readers coming from OOP)

| Java | Go |
|------|-----|
| **Class** = blueprint (fields + methods) | **Struct** = data only (fields). Methods are attached separately. |
| `new MyClass()` creates an object | `&MyStruct{}` or `MyStruct{}` creates a value |
| Inheritance (`extends`) | No inheritance. Use **embedding** instead (one struct inside another). |
| Class = single definition of type + behaviour | Struct = data. Interface = behaviour contract. |

In Java, a class defines both "what it is" and "what it can do." In Go, you separate: **structs** hold data, **interfaces** describe what a type can do (its methods). A struct doesn't "implement" an interface by declaration — it does so implicitly by having the right methods.

### Purpose of Interfaces

1. **Swap implementations without changing code** — A function that accepts `Client` works with the real Kubernetes client, a fake for tests, or a mock. The function stays the same; you inject different implementations.
2. **Easier testing** — Instead of a real database, HTTP client, or Kubernetes API, pass a fake that satisfies the same interface. No need for the real dependency.
3. **Decouple packages** — Your controller depends on `client.Client` (the interface), not the concrete type from controller-runtime. It only needs "something that can Get and List." The caller chooses what to inject.
4. **Polymorphism** — Write functions that work with "any type that can do X" without knowing the concrete type. For example, `io.Copy(dst Writer, src Reader)` works with files, buffers, network connections, or strings — anything that implements `Reader` and `Writer`. Same code, different implementations.
5. **Cross-package contracts** — Define an interface in your package; other packages implement it without you importing them. Dependencies point inward toward your code, not outward.
6. **Pluggable behaviour** — Strategies, HTTP handlers, middlewares: swap implementations at runtime based on config or context.
7. **Built-in types as interfaces** — In Go, `error` is an interface: `type error interface { Error() string }`. Any type with an `Error() string` method can be returned as an error. No inheritance required.

**Go proverb**: "Accept interfaces, return structs" — design function parameters around interfaces so callers can pass whatever they want; return concrete types.

### What "Implicit Implementation" Means

In languages like Java or C#, you explicitly declare:
```java
class MyClient implements Client { ... }
```

In Go, you declare **nothing**. If your type has the right methods, it automatically satisfies the interface:

```go
type Client interface {
    Get(ctx context.Context, key ObjectKey, obj Object) error
    List(ctx context.Context, list ObjectList, opts ...ListOption) error
}

// Some concrete type (you never see its definition — it lives in controller-runtime)
// It has Get(...) and List(...) with exactly those signatures.
// Therefore, it satisfies Client. No "implements Client" needed.
```

The compiler checks: "Does this type have a `Get` method with that signature? A `List` method with that signature?" If yes, the type can be used wherever `Client` is expected.

### How the Compiler Decides

For a type `T` to satisfy interface `I`:

1. `I` lists methods: `M1(...)`, `M2(...)`, etc.
2. `T` must have methods with the **exact same names** and **exact same signatures** (parameters and return types).
3. If `T` has pointer receiver methods, use `*T`, not `T`, when assigning to the interface.

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

// This type satisfies Reader — it has Read with the right signature
type FileReader struct { ... }
func (f *FileReader) Read(p []byte) (int, error) { ... }

// This also satisfies Reader — different struct, same method contract
type BufferReader struct { ... }
func (b *BufferReader) Read(p []byte) (int, error) { ... }

// Both can be passed to a function that expects Reader
func Process(r Reader) { ... }
Process(&FileReader{})   // works
Process(&BufferReader{}) // works
```

### Why This Matters: Loose Coupling

Functions (or structs) that depend on an **interface** instead of a **concrete type** can work with any implementation:

```go
type KafkaTopicReconciler struct {
    client.Client   // interface, not a concrete *Client
    Scheme *runtime.Scheme
}
```

The reconciler only needs something that can `Get` and `List` Kubernetes objects. It doesn't care whether that's:

- The real Kubernetes API server client (in production)
- A fake client for unit tests
- A cached client

As long as the value has `Get` and `List` with the right signatures, it satisfies `Client`. You inject the real `k8sClient` when running, or a fake when testing. The reconciler code stays the same.

### In This Codebase

- `KafkaTopicReconciler` embeds `client.Client` — it expects *any* value that can `Get`/`List`/etc. Kubernetes objects.
- `k8sClient` (from `client.New(...)`) is a concrete type that has those methods. It satisfies `Client`.
- In tests: `KafkaTopicReconciler{Client: k8sClient, Scheme: scheme}` — you pass the real envtest client. In other tests, you could pass a fake. The reconciler doesn't need to change.

### Summary

| Concept | Meaning |
|---------|---------|
| Interface | Contract = list of method signatures |
| Implement (implicitly) | Your type has methods matching those signatures |
| No `implements` keyword | Compiler infers it; no explicit declaration needed |
| Purposes | Testing/mocks, decoupling, polymorphism, cross-package contracts, pluggable behaviour |
| Go proverb | "Accept interfaces, return structs" — parameters as interfaces, return concrete types |

For a detailed explanation of how interfaces work at runtime (e.g. why `r.Get()` executes real code when the interface only defines signatures), see **[interfaces-runtime-explained.md](interfaces-runtime-explained.md)**.

---

## 14. The Blank Identifier `_`

Besides "import for side effects," `_` is used to discard values you don't need:

```go
result, _ := returnsTwoValues()  // ignore the second return value
```

Use sparingly — often you should handle the error instead of discarding it.

---

## Quick Reference

| Concept        | In One Line                                                                 |
|----------------|-----------------------------------------------------------------------------|
| Package        | Directory of Go files sharing a name; `main` for executables.              |
| Import         | Bring in other packages; `_` for side effects only.                        |
| Variable       | `var x T` or `x := value`. Package-level vars are shared.                  |
| Function       | `func name(args) returnType { }`. Multiple returns: `(value, error)`.       |
| Error handling | Check `err`; no exceptions.                                                |
| `init()`       | Runs when package loads, before `main()`.                                  |
| Struct         | Groups fields. Spec/Status in CRDs are structs.                             |
| Struct tags    | `` `json:"name"` `` (no space); omitempty/omitzero for optional; Kubebuilder markers in comments above fields. |
| Embedding      | One struct embeds another to reuse its fields.                              |
| Pointer        | `*T` references a value; `&x` takes address.                              |
| Method         | `func (r *T) Name()` — function on a type.                                  |
| Context        | Carries deadlines and cancellation.                                       |
| Interface      | Contract of methods; satisfied implicitly.                                |

---

## Related Docs

- **[controller-reconcile-concepts.md](controller-reconcile-concepts.md)** — Go concepts specific to the controller and Reconcile function (embedding, `ctrl.Request`, method chaining, client usage).
- **[test-concepts.md](test-concepts.md)** — Test concepts (Ginkgo, Gomega, envtest) and Go patterns in controller tests.

---

## Further Reading

- [A Tour of Go](https://go.dev/tour/) — official interactive tutorial
- [Effective Go](https://go.dev/doc/effective_go) — style and idioms
- [Go by Example](https://gobyexample.com/) — short examples by topic
