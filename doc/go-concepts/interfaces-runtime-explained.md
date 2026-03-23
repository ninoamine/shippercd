# How Interfaces Work at Runtime — A Detailed Explanation

**For**: Beginners confused about "the interface only defines the signature, so how does the code actually run?"

---

## The Confusion

You see this:

```go
type KafkaTopicReconciler struct {
    client.Client   // interface — defines Get(), List(), etc. with only signatures
    Scheme *runtime.Scheme
}

func (r *KafkaTopicReconciler) Reconcile(...) {
    err := r.Get(ctx, req.NamespacedName, &kafkaTopic)  // ← How does this work?
}
```

The interface `client.Client` only declares *what* methods exist (names, parameters, return types). It does **not** contain the actual code that talks to Kubernetes. So when you call `r.Get()`, where does the real behavior come from?

---

## The Key Idea: Two Things at Once

An interface value in Go holds **two** pieces of information:

1. **The type** of the concrete value (e.g. "this is a *client.Client from controller-runtime")
2. **A pointer to the concrete value** (the real object in memory)

When you call a method through the interface, Go uses that stored concrete value to execute the method. The interface is just a "window" — it lets you call methods, but the **implementation** lives in the concrete type.

---

## Step-by-Step: Who Provides the Implementation?

### 1. The Interface Is a Contract (No Code)

```go
// Simplified view of what client.Client looks like
type Client interface {
    Get(ctx context.Context, key ObjectKey, obj Object) error
    List(ctx context.Context, list ObjectList, opts ...ListOption) error
    Create(ctx context.Context, obj Object, opts ...CreateOption) error
    // ... more methods
}
```

This says: "To satisfy `Client`, a type must have methods named `Get`, `List`, `Create`, etc. with these exact signatures."  
It does **not** say what those methods do. It's a contract only.

---

### 2. A Concrete Type Implements the Contract

Somewhere in the controller-runtime package, there is a **concrete type** (a struct) that has real code:

```go
// This lives in controller-runtime — you don't see it, but it exists
type client struct {
    scheme *runtime.Scheme
    cache  runtime.Cache
    // ... other fields
}

func (c *client) Get(ctx context.Context, key ObjectKey, obj Object) error {
    // REAL CODE: makes HTTP request to Kubernetes API, fills obj
    return c.reader.Get(ctx, key, obj)
}

func (c *client) List(ctx context.Context, list ObjectList, opts ...ListOption) error {
    // REAL CODE: lists resources from the cluster
    // ...
}
```

This struct has the **actual implementation**. The `Get` method does the real work: it talks to the Kubernetes API, fetches the resource, and fills the `obj` you passed in.

---

### 3. You Inject the Concrete Value When Creating the Reconciler

When the manager starts your controller, it does something like this (simplified):

```go
// In cmd/main.go or wherever the manager is set up
mgr, _ := ctrl.NewManager(cfg, ctrl.Options{...})

// The manager creates a real Kubernetes client
k8sClient, _ := client.New(cfg, client.Options{Scheme: scheme.Scheme()})
// k8sClient is a *client (the concrete type from step 2)
// It has real Get(), List(), etc. that talk to the API server

// When registering the reconciler
reconciler := &KafkaTopicReconciler{
    Client: k8sClient,   // ← You put the CONCRETE value here
    Scheme: scheme.Scheme(),
}

mgr.Add reconciler...
```

At this moment, `reconciler.Client` is an interface variable that **holds** the concrete `k8sClient`. The interface doesn't replace it — it wraps it. The concrete value is stored inside.

---

### 4. When You Call `r.Get()`, Go Uses the Stored Concrete Value

```go
err := r.Get(ctx, req.NamespacedName, &kafkaTopic)
```

What happens:

1. `r` is the `KafkaTopicReconciler`
2. `r.Client` is an interface value containing the concrete `k8sClient`
3. `r.Get` is actually `r.Client.Get` (because `Client` is embedded, so `r` inherits its methods)
4. Go looks at what's inside the interface: "Ah, it's a `*client` from controller-runtime"
5. Go calls the `Get` method **on that concrete value**
6. The real `Get` implementation runs — it talks to Kubernetes and fills `kafkaTopic`

The interface doesn't "do" the work. It's a gate that says "whatever is here must have `Get`." The thing behind the gate (the concrete `k8sClient`) has the real code, and that's what runs.

---

## Visual Summary

```
┌─────────────────────────────────────────────────────────────────────┐
│  KafkaTopicReconciler                                                │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐  │
│  │  Client (field type: client.Client = INTERFACE)                 │  │
│  │                                                                 │  │
│  │  Interface value contains:                                      │  │
│  │  ┌──────────────────────────────────────────────────────────┐   │  │
│  │  │  (type)  →  *client (from controller-runtime)            │   │  │
│  │  │  (value) →  pointer to real k8sClient in memory            │   │  │
│  │  └──────────────────────────────────────────────────────────┘   │  │
│  │                                                                 │  │
│  │  That *client has:                                               │  │
│  │    Get()  → real code that calls Kubernetes API                  │  │
│  │    List() → real code that lists resources                       │  │
│  │    ...                                                           │  │
│  └────────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  When you call r.Get():                                               │
│    1. r.Get resolves to r.Client.Get (embedding)                      │
│    2. Go looks inside r.Client → finds *client                        │
│    3. Go calls (*client).Get(...) → the REAL implementation runs     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Simpler Example: Reader

Same idea with a smaller example:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

// FileReader has REAL code
type FileReader struct { f *os.File }
func (f *FileReader) Read(p []byte) (int, error) {
    return f.f.Read(p)  // Actually reads from disk
}

// BufferReader has DIFFERENT real code
type BufferReader struct { buf []byte }
func (b *BufferReader) Read(p []byte) (int, error) {
    n := copy(p, b.buf)  // Actually copies from buffer
    return n, nil
}

func Process(r Reader) {
    data := make([]byte, 100)
    r.Read(data)  // Which Read runs? Whatever concrete type was passed!
}

Process(&FileReader{...})   // FileReader.Read runs
Process(&BufferReader{...})  // BufferReader.Read runs
```

`Process` doesn't know if it got a file or a buffer. It just calls `r.Read()`. Go looks at what's actually in `r` and runs that type's `Read` method. The interface is the contract; the concrete type has the code.

---

## In Tests: A Different Concrete Value

In your test:

```go
k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme()})
// This creates a client that talks to envtest (fake API server in memory)

controllerReconciler := &KafkaTopicReconciler{
    Client: k8sClient,   // Same interface, different "kind" of implementation
    Scheme: k8sClient.Scheme(),
}
```

Here, `k8sClient` is still a concrete type that implements `Client`. It might be the same `*client` struct, but configured to talk to envtest instead of a real cluster. Or it could be a different concrete type (e.g. a fake). Either way, it has real `Get`/`List` code — they just do something different (talk to envtest instead of a real API).

The reconciler code is identical. It only cares: "give me something with `Get` and `List`." You give it whatever suits the situation.

---

## Summary

| Question | Answer |
|----------|--------|
| Where is the implementation? | In the **concrete type** (e.g. `*client` in controller-runtime). |
| What does the interface do? | Declares the contract (method signatures). At runtime, it holds a reference to the concrete value. |
| When I call `r.Get()`, what runs? | The `Get` method of the **concrete value** stored in `r.Client`. |
| How does the reconciler get that concrete value? | It's **injected** when the reconciler is created (in `main.go` or tests). |

**Bottom line**: The interface is a slot. You put a concrete value in the slot when you create the struct. When you call a method through the interface, Go runs the method on whatever concrete value is in the slot. The interface never contains code — it only describes the shape of what can go in the slot.

---

## Related Docs

- **[concepts.md](concepts.md)** — Section 13 (Interfaces) for the basics
- **[controller-reconcile-concepts.md](controller-reconcile-concepts.md)** — How the reconciler uses the client
