# Test Concepts and Go Patterns in Controller Tests

**For**: Beginners in Go and Kubernetes operator testing  
**Purpose**: Explains the testing framework, patterns, and Go concepts used in the KafkaTopic controller tests.

This doc covers both **testing concepts** (Ginkgo, Gomega, envtest) and **Go language patterns** you will see in the test files.

---

## Overview

Kubebuilder scaffolds controller tests that use:

1. **Ginkgo** — A BDD-style (Behavior-Driven Development) testing framework
2. **Gomega** — A matcher/assertion library used with Ginkgo
3. **envtest** — A test environment that runs a real Kubernetes API server (and etcd) in-process — no real cluster needed

The tests live in two files:

- **`suite_test.go`** — Sets up and tears down the test environment; defines shared variables
- **`*_controller_test.go`** — The actual test cases (e.g. `kafkatopic_controller_test.go`)

---

## Part 1: Testing Concepts

### Ginkgo: BDD-Style Structure

Ginkgo organizes tests in a nested, descriptive structure:

| Term | Purpose |
|------|---------|
| **`Describe`** | Groups related tests; describes what you are testing (e.g. "KafkaTopic Controller") |
| **`Context`** | Sub-groups; describes a scenario (e.g. "When reconciling a resource") |
| **`It`** | A single test case; describes the expected behavior (e.g. "should successfully reconcile the resource") |

Each takes a string and a function. The function contains the test logic or more nested blocks.

```go
var _ = Describe("KafkaTopic Controller", func() {
    Context("When reconciling a resource", func() {
        It("should successfully reconcile the resource", func() {
            // test logic here
        })
    })
})
```

**`var _ =`** — Runs the block at package init time. Ginkgo uses this to register the test structure. The `_` discards the return value.

---

### Ginkgo: Lifecycle Hooks

| Hook | When It Runs |
|------|--------------|
| **`BeforeSuite`** | Once, before any tests in the suite |
| **`AfterSuite`** | Once, after all tests finish |
| **`BeforeEach`** | Before each `It` in the same `Describe`/`Context` |
| **`AfterEach`** | After each `It` in the same `Describe`/`Context` |

Use `BeforeEach` to create test resources; use `AfterEach` to clean them up. That way each test starts from a known state.

---

### Ginkgo: `By()` for Readable Output

`By("description")` adds a step description to the test output. When a test fails, you see which step was running:

```go
By("creating the custom resource for the Kind KafkaTopic")
By("Reconciling the created resource")
By("Verifying the status condition is set")
```

---

### Gomega: Assertions with `Expect`

Gomega provides fluent assertions:

```go
Expect(err).NotTo(HaveOccurred())
Expect(k8sClient.Create(ctx, resource)).To(Succeed())
Expect(readyCondition).NotTo(BeNil())
Expect(readyCondition.Status).To(Equal(metav1.ConditionTrue))
```

| Matcher | Meaning |
|---------|---------|
| `To(Succeed())` | The value is `nil` (no error) |
| `NotTo(HaveOccurred())` | The value (usually `err`) is `nil` |
| `NotTo(BeNil())` | The value is not `nil` |
| `To(Equal(x))` | The value equals `x` |
| `To(BeNil())` | The value is `nil` |

If a matcher fails, Ginkgo marks the test as failed and prints a helpful message.

---

### Gomega: `Eventually` for Async Checks

`Eventually` retries a function until it succeeds or times out:

```go
Eventually(func() error {
    return testEnv.Stop()
}, time.Minute, time.Second).Should(Succeed())
```

- First argument: a function that returns an error
- Second: timeout (e.g. 1 minute)
- Third: polling interval (e.g. 1 second)
- `.Should(Succeed())` — The function must eventually return `nil`

Use `Eventually` when you need to wait for something (e.g. teardown, eventual consistency).

---

### envtest: In-Process Kubernetes API

**envtest** starts a real Kubernetes API server and etcd in the same process as your tests. No Docker, no Kind, no cluster — just `go test`.

In `suite_test.go`:

```go
testEnv = &envtest.Environment{
    CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
    ErrorIfCRDPathMissing: true,
}
cfg, err = testEnv.Start()
k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
```

- **`CRDDirectoryPaths`** — Where to find your CRD YAML files. envtest installs them so the API knows about KafkaTopic, Environment, etc.
- **`testEnv.Start()`** — Starts the API server; returns a `*rest.Config` (kubeconfig)
- **`client.New(cfg, ...)`** — Creates a controller-runtime client that talks to this in-process API

`k8sClient` is shared across tests. You use it to Create, Get, Delete resources — just like in a real cluster.

---

### Test Entry Point

```go
func TestControllers(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Controller Suite")
}
```

- **`TestControllers`** — The function that `go test` calls. Must take `*testing.T`.
- **`RegisterFailHandler(Fail)`** — Tells Ginkgo how to fail a test (call `Fail()` which integrates with `testing.T`).
- **`RunSpecs(t, "Controller Suite")`** — Runs all registered `Describe`/`Context`/`It` blocks.

---

## Part 2: Go Concepts in the Tests

### Dot Import

```go
. "github.com/onsi/ginkgo/v2"
. "github.com/onsi/gomega"
```

The **dot** (`.`) imports the package into the current namespace. So you write `Describe` instead of `ginkgo.Describe`, and `Expect` instead of `gomega.Expect`. Common in test files for readability; avoid in production code.

---

### Package-Level Variables

```go
var (
    ctx       context.Context
    cancel    context.CancelFunc
    testEnv   *envtest.Environment
    cfg       *rest.Config
    k8sClient client.Client
)
```

These are **package-level** — shared by all tests in the package. `BeforeSuite` sets them; `BeforeEach` and `It` blocks use them. They persist for the whole test run.

---

### Constants and Variables in `Describe`/`Context`

```go
Context("When reconciling a resource", func() {
    const resourceName = "test-resource"
    ctx := context.Background()
    typeNamespacedName := types.NamespacedName{Name: resourceName, Namespace: "default"}
    kafkatopic := &kafkav1alpha1.KafkaTopic{}
    // ...
})
```

- **`const`** — Immutable; good for fixed test values
- **`:=`** — Short variable declaration; type inferred
- These are **closure variables** — the inner `BeforeEach`, `AfterEach`, and `It` blocks capture them and can use them

---

### Anonymous Functions (Closures)

Ginkgo blocks are **anonymous functions** — functions without a name:

```go
It("should successfully reconcile the resource", func() {
    // this function runs when the test executes
})
```

The inner function can access variables from the outer scope (e.g. `ctx`, `kafkatopic`, `typeNamespacedName`). That is a **closure** — the function "closes over" those variables.

---

### Composite Literals

```go
resource := &kafkav1alpha1.KafkaTopic{
    ObjectMeta: metav1.ObjectMeta{
        Name:      resourceName,
        Namespace: "default",
    },
}
```

A **composite literal** constructs a struct by listing its fields. `&` creates a pointer to it. `ObjectMeta` is nested — another composite literal inside.

---

### The Blank Identifier `_`

```go
_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{...})
```

`_` discards the first return value (`ctrl.Result`). The test only cares about `err`.

---

### Pointer to Empty Struct

```go
controllerReconciler := &KafkaTopicReconciler{
    Client: k8sClient,
    Scheme: k8sClient.Scheme(),
}
```

Creates a reconciler with the test client and scheme. `k8sClient` and `k8sClient.Scheme()` come from the envtest setup in `BeforeSuite`.

---

### Conditional Create (Idempotent Setup)

```go
BeforeEach(func() {
    err := k8sClient.Get(ctx, typeNamespacedName, kafkatopic)
    if err != nil && errors.IsNotFound(err) {
        resource := &kafkav1alpha1.KafkaTopic{...}
        Expect(k8sClient.Create(ctx, resource)).To(Succeed())
    }
})
```

Only creates the resource if it does not exist. Useful when tests share state or when `BeforeEach` runs multiple times (e.g. in nested contexts).

---

## Part 3: Test Flow Summary

```
BeforeSuite
  ├─ Set up logger
  ├─ Start envtest (API server + etcd)
  ├─ Install CRDs
  └─ Create k8sClient

For each It:
  BeforeEach
    └─ Create KafkaTopic if not found
  It("should successfully reconcile...")
    ├─ Build reconciler with k8sClient
    ├─ Call Reconcile
    ├─ Assert no error
    └─ (Optional) Assert status condition
  AfterEach
    └─ Delete the KafkaTopic

AfterSuite
  └─ Stop envtest
```

---

## Quick Reference: Test Concepts

| Concept | Where You See It |
|---------|------------------|
| Describe | Top-level test group |
| Context | Scenario sub-group |
| It | Single test case |
| BeforeEach / AfterEach | Setup and cleanup per test |
| By() | Step description in output |
| Expect().To() / NotTo() | Gomega assertions |
| Eventually | Async retry with timeout |
| envtest | In-process Kubernetes API |
| k8sClient | Client for Create/Get/Delete |
| Dot import | `Describe`, `Expect` without package prefix |
| Closure | Inner blocks access outer variables |
| Package-level var | Shared testEnv, k8sClient, cfg |

---

## Further Reading

- [Ginkgo documentation](https://onsi.github.io/ginkgo/)
- [Gomega matchers](https://onsi.github.io/gomega/)
- [controller-runtime envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest)
- [Kubebuilder Book: Testing](https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html)
