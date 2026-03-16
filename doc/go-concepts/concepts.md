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

A **pointer** holds the address of a value. `*T` is "pointer to T".

- `&x` — address of `x`
- `*p` — value pointed to by `p`
- `nil` — zero value for pointers

**Why use pointers?**
- Avoid copying large structs
- Allow a function to mutate the original value
- Represent "optional" (nil = absent)

---

## 11. Methods and Receivers

A **method** is a function attached to a type:

```go
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ...
}
```

- `(r *Reconciler)` — the **receiver**. The method is called on a `*Reconciler` value.
- **Pointer receiver** `*Reconciler` — the method can modify the receiver and avoids copying.
- **Value receiver** `Reconciler` — the method receives a copy; use for small types you don't need to mutate.

---

## 12. Context

`context.Context` is a standard way to carry deadlines, cancellation signals, and request-scoped values. It is often the first parameter of long-running or cancelable operations.

```go
func Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
```

When the context is cancelled (e.g. shutting down), code should stop work and return. Controller-runtime passes a context into `Reconcile`.

---

## 13. Interfaces

An **interface** describes behaviour — a set of methods. Any type that implements those methods satisfies the interface implicitly (no `implements` keyword).

```go
type Client interface {
    Get(ctx context.Context, key ObjectKey, obj Object) error
    List(ctx context.Context, list ObjectList, opts ...ListOption) error
}
```

If your type has `Get` and `List` with the right signatures, it satisfies `Client`. Interfaces enable loose coupling and testability.

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
