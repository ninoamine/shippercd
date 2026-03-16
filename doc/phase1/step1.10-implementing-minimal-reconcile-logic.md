# Phase 1, Step 1.10: Implementing Minimal Reconcile Logic

**For**: Beginners in Go and Kubernetes  
**Goal**: Understand the reconcile loop, how to log from it, and how to update Status — so you can implement it yourself without copying code.

---

## What You Are Doing

In this step, you will add logic to the KafkaTopic controller's `Reconcile` function. The goal is minimal: make the controller **react** when a KafkaTopic is created or updated, and optionally **update its Status**. You will not call Kafka yet — that comes in later phases.

By the end, when you apply a KafkaTopic resource and run the controller, you should see log output and (if you choose) an updated status field when you run `kubectl get kafkatopic -o yaml`.

---

## Where to Work

The KafkaTopic controller lives in one of these locations, depending on your project layout:

| Layout | Controller path |
|--------|------------------|
| **Multi-group** | `internal/controller/kafka/kafkatopic_controller.go` |
| **Single-group** (KafkaTopic in core) | `internal/controller/kafkatopic_controller.go` |

Open that file and find the `Reconcile` function. It receives a `ctrl.Request` (which has `Namespace` and `Name`) and returns `(ctrl.Result, error)`.

---

## Concepts: The Reconcile Loop

The reconcile loop is the heart of a Kubernetes controller. It runs whenever:

- A KafkaTopic is created, updated, or deleted
- The controller starts (it reconciles all existing KafkaTopics)
- You explicitly requeue a request

Your job is to make the **actual cluster state** match the **desired state** (the Spec). For Phase 1, "actual state" means: the controller noticed the resource and optionally wrote something to Status.

---

## Step-by-Step Walkthrough

Follow these steps in order. Each step tells you exactly what to change.

---

### Step 1: Fix the logger (required)

**Current line** (around line 50):
```go
_ = logf.FromContext(ctx)
```

**Problem**: The `_` discards the logger. You cannot call methods on it.

**Change**: Replace `_` with a variable name, e.g. `log`:
```go
log := logf.FromContext(ctx)
```

Now you have a logger you can use.

---

### Step 2: Add a log line (required)

**Where**: Right after the logger line, before the `return`.

**What to add**: A call to `log.Info(...)`. The `Info` method takes:
1. A message string (e.g. `"Reconciling KafkaTopic"`)
2. Alternating key-value pairs: `"name", value, "namespace", value`

**How to get the values**: The `req` parameter has `req.NamespacedName`, which has `.Name` and `.Namespace`. So:
- For the name: `req.NamespacedName.Name`
- For the namespace: `req.NamespacedName.Namespace`

**Structure**:
```go
log.Info("Your message here", "name", req.NamespacedName.Name, "namespace", req.NamespacedName.Namespace)
```

Replace `"Your message here"` with something like `"Reconciling KafkaTopic"` (capital letter, no period at the end).

---

### Step 3: Fetch the KafkaTopic (needed for status update)

**Where**: After logging, before the return.

**Add an import** (in the `import` block): `"k8s.io/apimachinery/pkg/api/errors"` — you need this to check if the resource was not found.

**Declare a variable** for the KafkaTopic:
```go
var kafkaTopic kafkav1alpha1.KafkaTopic
```

**Call Get** to fetch it from the cluster:
```go
err := r.Get(ctx, req.NamespacedName, &kafkaTopic)
```

**Handle the error**: If `Get` fails, the resource might be deleted. Check with `errors.IsNotFound(err)`. If true, return `ctrl.Result{}, nil` (nothing to reconcile). Otherwise, return `ctrl.Result{}, err` (let the controller retry).

**Structure**:
```go
err := r.Get(ctx, req.NamespacedName, &kafkaTopic)
if err != nil {
    if errors.IsNotFound(err) {
        return ctrl.Result{}, nil
    }
    return ctrl.Result{}, err
}
```

---

### Step 4: Update the Status (optional but recommended)

**Add imports**: `"k8s.io/apimachinery/pkg/api/meta"` and `"k8s.io/apimachinery/pkg/apis/meta/v1"` (you may already have `metav1` — that is `apis/meta/v1`).

**Add a condition** to `kafkaTopic.Status.Conditions` using `meta.SetStatusCondition`. The function signature is:
```go
meta.SetStatusCondition(&slice, condition)
```

You pass a pointer to the conditions slice and a `metav1.Condition`. To build a condition:
- `Type`: `"Ready"` (string)
- `Status`: `metav1.ConditionTrue` (use this constant)
- `Reason`: `"Reconciled"` (string)
- `Message`: `"Reconciled successfully"` (string)

`metav1.Condition` has other fields (e.g. `LastTransitionTime`); you can set them or leave them. The helper `meta.SetStatusCondition` will set `LastTransitionTime` if needed.

**Then update the status** in the cluster:
```go
err = r.Status().Update(ctx, &kafkaTopic)
```

Note: `Status()` not `Update` directly. And `err =` (assign, not `:=`) because `err` was already declared.

**Handle the error**: If `Status().Update` fails, return `ctrl.Result{}, err` so the controller retries.

---

### Step 5: Return success

Keep the existing `return ctrl.Result{}, nil` at the end. Your logic runs before it.

---

## Code Skeleton (fill in the blanks)

Here is the structure of the `Reconcile` function. Use it as a guide for the order of operations:

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

    // Add a Ready condition
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

**Imports you need** (add any that are missing):
- `"k8s.io/apimachinery/pkg/api/errors"`
- `"k8s.io/apimachinery/pkg/api/meta"`
- `metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"`

---

## Minimal Path (logging only)

If you want to start with just logging and skip the status update:

1. Do Step 1 and Step 2 above.
2. Return `ctrl.Result{}, nil`.
3. Skip Steps 3 and 4.

That is enough to see your log line when the controller runs. You can add the fetch and status update later.

---

## What to Return

For success with no retry:

- Return `ctrl.Result{}, nil`

For "reconcile again after a delay" (e.g. to poll):

- Return `ctrl.Result{RequeueAfter: 5 * time.Minute}, nil`

For "something went wrong, retry":

- Return `ctrl.Result{}, err`

For Phase 1, returning `ctrl.Result{}, nil` after your logging (and optional status update) is sufficient.

---

📖 **Go concepts**: See [Controller and Reconcile Concepts](../go-concepts/controller-reconcile-concepts.md) for explanations of struct embedding, `ctrl.Request`/`ctrl.Result`, method chaining, client usage, and other patterns used in the controller code.

---

## Packages You Will Likely Use

| Package | Purpose |
|---------|---------|
| `sigs.k8s.io/controller-runtime/pkg/log` | Logger from context |
| `sigs.k8s.io/controller-runtime/pkg/client` | Already available via `r.Client` |
| `k8s.io/apimachinery/pkg/types` | `NamespacedName` for Get |
| `k8s.io/apimachinery/pkg/api/meta` | `SetStatusCondition` for conditions |
| `k8s.io/apimachinery/pkg/apis/meta/v1` | `metav1.Condition`, `metav1.Now()` |

---

## Checklist: Do It Yourself

- [ ] Open the KafkaTopic controller file and locate the `Reconcile` function
- [ ] Obtain a logger from the context and log a message when Reconcile runs, including name and namespace
- [ ] (Optional) Fetch the KafkaTopic using `r.Get` and a `NamespacedName`
- [ ] (Optional) Update Status — either add a condition or a message field
- [ ] Return `ctrl.Result{}, nil` on success
- [ ] Run `make generate` and `make manifests` if you changed the types
- [ ] Run `make test` to ensure tests pass
- [ ] Run `make run`, apply a KafkaTopic, and verify you see your log line (and status if you added it)

---

## What Can Go Wrong?

- **Logger is discarded**: The scaffold may assign the logger to `_`. Replace that with a variable so you can call `Info` on it.
- **Status update fails with conflict**: You may be updating a stale object. Re-fetch right before the update, or use `Patch` instead of `Update`.
- **Condition not showing in YAML**: Ensure you call `Status().Update` (or `Patch`), not `Update`. The Status subresource is separate from the main resource.
- **`SetStatusCondition` not found**: It lives in `k8s.io/apimachinery/pkg/api/meta`. Add the import.

---

## Next Step

In **Step 1.11**, you will run `make install` to install the CRDs into your cluster, then `make run` to start the controller. In Step 1.13 you will create a sample KafkaTopic and apply it. Your new logging and status updates will appear when the controller reconciles that resource.
