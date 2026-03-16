# Go Concepts in the Controller and Reconcile Function

**For**: Beginners in Go working on Kubebuilder controllers  
**Purpose**: Explains the Go concepts and patterns used in controller code and the `Reconcile` function.

This doc complements [concepts.md](concepts.md) by focusing on controller-specific code. Read concepts.md first for fundamentals (structs, pointers, methods, context, interfaces).

---

## Overview

A Kubebuilder controller has two main parts:

1. **The reconciler struct** — Holds dependencies (client, scheme) and implements `Reconcile`
2. **The `Reconcile` function** — Called whenever a watched resource changes; your logic lives here

---

## 1. Struct Embedding in the Reconciler

The reconciler embeds other types to get their methods:

```go
type KafkaTopicReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}
```

**Embedding** means `KafkaTopicReconciler` gets all methods of `client.Client` as if they were defined on it. So you can call `r.Get(...)` and `r.Status().Update(...)` directly on `r` — the embedded `Client` provides those methods.

- **`client.Client`** — Interface for talking to the Kubernetes API (Get, List, Create, Update, Patch, Delete, etc.)
- **`Scheme`** — Knows how to serialize/deserialize your custom types; needed when creating or updating objects

You do not need to write `r.Client.Get(...)` — embedding promotes the methods to the outer struct.

---

## 2. Pointer Receiver for Reconcile

```go
func (r *KafkaTopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
```

The receiver `(r *KafkaTopicReconciler)` is a **pointer receiver**:

- `r` refers to the same reconciler instance that was registered with the manager
- The method can use `r.Client` and `r.Scheme` (the embedded fields)
- A pointer avoids copying the struct; the reconciler may hold references to shared resources

---

## 3. The Reconcile Function Signature

| Parameter / Return | Type | Purpose |
|--------------------|------|---------|
| `ctx` | `context.Context` | Carries deadlines and cancellation; pass it to all API calls |
| `req` | `ctrl.Request` | Identifies the resource to reconcile (`Namespace`, `Name`) |
| Return 1 | `ctrl.Result` | Tells the controller whether to requeue and when |
| Return 2 | `error` | Non-nil triggers a retry; nil means success |

**`ctrl.Request`** wraps `types.NamespacedName`. Use `req.NamespacedName` to fetch the object: `r.Get(ctx, req.NamespacedName, &obj)`.

**`ctrl.Result`** has:

- `Requeue bool` — Requeue immediately
- `RequeueAfter time.Duration` — Requeue after a delay (e.g. for polling)

An empty result `ctrl.Result{}` means "done, no requeue."

---

## 4. Multiple Return Values

`Reconcile` returns two values: `(ctrl.Result, error)`. Go functions often return `(value, error)` — the caller checks the error first.

When you return `ctrl.Result{}, nil`, you are returning:
- An empty struct (zero value for `ctrl.Result`)
- `nil` for the error (no failure)

---

## 5. Composite Literals

A **composite literal** constructs a value by listing its fields:

```go
return ctrl.Result{}, nil
```

`ctrl.Result{}` is a composite literal with no fields set — all fields get their zero values (`Requeue: false`, `RequeueAfter: 0`).

To requeue after 5 minutes:

```go
return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
```

---

## 6. Import Aliases

Controller code uses short aliases for long package paths:

| Alias | Package | Why |
|-------|---------|-----|
| `ctrl` | `sigs.k8s.io/controller-runtime` | Shorter than typing the full path; `ctrl.Request`, `ctrl.Result`, `ctrl.NewControllerManagedBy` |
| `logf` | `sigs.k8s.io/controller-runtime/pkg/log` | Logger package; `logf.FromContext(ctx)` |
| `kafkav1alpha1` | `github.com/.../api/kafka/v1alpha1` | Avoids name clash with `v1alpha1` from another group |

---

## 7. Method Chaining (Builder Pattern)

`SetupWithManager` uses a **builder pattern** — each method returns the builder so you can chain:

```go
return ctrl.NewControllerManagedBy(mgr).
    For(&kafkav1alpha1.KafkaTopic{}).
    Named("kafka-kafkatopic").
    Complete(r)
```

- **`NewControllerManagedBy(mgr)`** — Start building a controller for this manager
- **`For(&kafkav1alpha1.KafkaTopic{})`** — Watch KafkaTopic resources; pass a pointer to the type (not an instance)
- **`Named("...")`** — Optional; sets the controller name for logging and metrics
- **`Complete(r)`** — Register the reconciler; `r` must implement the `Reconciler` interface

The `&kafkav1alpha1.KafkaTopic{}` is a pointer to an empty KafkaTopic — used only for its type, not its data.

---

## 8. The Reconciler Interface

Controller-runtime expects your reconciler to implement:

```go
type Reconciler interface {
    Reconcile(context.Context, Request) (Result, error)
}
```

If your struct has a method `Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)`, it satisfies this interface. No explicit `implements` keyword — Go uses **structural typing**: if the method signature matches, the type implements the interface.

---

## 9. Logging from Context

```go
log := logf.FromContext(ctx)
log.Info("Reconciling KafkaTopic", "name", req.NamespacedName.Name, "namespace", req.NamespacedName.Namespace)
```

Use `:=` to declare the logger as a local variable. Do not use `log =` — that would assign to a package-level variable that does not exist.

`logf.FromContext(ctx)` returns a logger that carries request-scoped data. The context is passed through the call chain, so logs from the same reconcile can be correlated.

**Structured logging**: `Info` takes a message and alternating key-value pairs. Each pair becomes a structured field (e.g. `"name": "my-topic"`).

---

## 10. Fetching Objects with the Client

To read a resource, you use the embedded client:

```go
var kafkaTopic kafkav1alpha1.KafkaTopic
err := r.Get(ctx, req.NamespacedName, &kafkaTopic)
if err != nil {
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, err
}
```

- **`req.NamespacedName`** — Use it directly; it has `Name` and `Namespace`
- **Third argument** — Pointer to the object; `Get` fills it with the data from the API server
- **Error handling** — If the resource was deleted, `Get` returns `errors.IsNotFound(err)`. Return `ctrl.Result{}, nil` (nothing to reconcile). Otherwise return the error so the controller retries

---

## 11. Updating Status (Subresource)

Status is a **subresource**. Use `r.Status().Update(ctx, &obj)` instead of `r.Update(ctx, &obj)`:

```go
meta.SetStatusCondition(&kafkaTopic.Status.Conditions, metav1.Condition{
	Type:    "Ready",
	Status:  metav1.ConditionTrue,
	Reason:  "Reconciled",
	Message: "Reconciled successfully",
})
if err := r.Status().Update(ctx, &kafkaTopic); err != nil {
	return ctrl.Result{}, err
}
```

- **`meta.SetStatusCondition`** — Adds or updates a condition in the slice; handles `LastTransitionTime` automatically
- **`r.Status()`** — Returns a subresource client; `Update` only touches status, not spec
- **`if err :=`** — Short declaration in the `if`; the `err` is scoped to the block

---

## 12. Controller Registration in main.go

In `cmd/main.go`, each reconciler is constructed and registered:

```go
if err := (&kafkacontroller.KafkaTopicReconciler{
    Client: mgr.GetClient(),
    Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
    // ...
}
```

- **`&kafkacontroller.KafkaTopicReconciler{...}`** — Composite literal; creates a pointer to a reconciler with fields set
- **`mgr.GetClient()`** — The manager's cached client (shared across controllers)
- **`mgr.GetScheme()`** — The scheme that knows your CRD types
- **`.SetupWithManager(mgr)`** — Method call on the struct; registers the controller with the manager

---

## 13. Test Patterns: types.NamespacedName and reconcile.Request

In controller tests you often see:

```go
typeNamespacedName := types.NamespacedName{
    Name:      resourceName,
    Namespace: "default",
}

_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
    NamespacedName: typeNamespacedName,
})
```

- **`types.NamespacedName`** — Same struct used by `client.Get`; identifies a resource
- **`reconcile.Request`** — Same as `ctrl.Request`; the `reconcile` package is the canonical one, `ctrl` re-exports it
- **`_`** — Blank identifier; discards the `ctrl.Result` because the test only cares about the error

---

## Quick Reference: Controller Code

| Concept | Where You See It |
|---------|-------------------|
| Struct embedding | `client.Client` and `Scheme` in the reconciler |
| Pointer receiver | `(r *KafkaTopicReconciler) Reconcile(...)` |
| Context | First parameter of `Reconcile`; pass to all API calls |
| ctrl.Request | `req.NamespacedName` (has `Name` and `Namespace`) |
| ctrl.Result | Return `ctrl.Result{}` for success; `RequeueAfter` for delayed retry |
| Multiple returns | `(ctrl.Result, error)` |
| Composite literal | `ctrl.Result{}`, `types.NamespacedName{Name: "x", Namespace: "y"}` |
| Import alias | `ctrl`, `logf`, `kafkav1alpha1` |
| Method chaining | `NewControllerManagedBy(mgr).For(...).Complete(r)` |
| Interface | Reconciler implements `Reconcile(context.Context, Request) (Result, error)` |
| Subresource client | `r.Status().Update(ctx, &obj)` for status-only updates |

---

## Full Reconcile Implementation (KafkaTopic)

Here is the complete `Reconcile` function from the KafkaTopic controller, with all concepts applied:

```go
func (r *KafkaTopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling KafkaTopic", "name", req.NamespacedName.Name, "namespace", req.NamespacedName.Namespace)

	var kafkaTopic kafkav1alpha1.KafkaTopic
	err := r.Get(ctx, req.NamespacedName, &kafkaTopic)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	meta.SetStatusCondition(&kafkaTopic.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "Reconciled",
		Message: "Reconciled successfully",
	})
	if err := r.Status().Update(ctx, &kafkaTopic); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
```

**Required imports** (in addition to the controller-runtime and API imports):

- `"k8s.io/apimachinery/pkg/api/errors"`
- `"k8s.io/apimachinery/pkg/api/meta"`
- `metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"`

---

## Further Reading

- [controller-runtime Reconciler](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile)
- [controller-runtime Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client)
- [Kubebuilder Book: Implementing a Controller](https://book.kubebuilder.io/cronjob-tutorial/controller-implementation.html)
