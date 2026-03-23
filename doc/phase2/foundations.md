# Phase 2: Environment Controller

**Estimated duration**: 1–2 weeks  
**Learning goal**: Controllers that create other CRs, ownership, finalizers. Use `Owns()` and `SetControllerReference()` for automatic child cleanup.

---

## Overview

Phase 2 turns the Environment controller into the **parent** that creates and manages child resources. You will:

- Scaffold PostgreSQL and MongoDB APIs (CRDs) if not already done
- Implement the Environment controller to watch `Environment` CRs and create/update `KafkaTopic`, `PostgresqlDatabase`, and `MongodbDatabase` CRs
- Use `ctrl.SetControllerReference()` so Kubernetes deletes child CRs when the parent Environment is deleted
- Add a finalizer to the Environment so the controller can clean up children before deletion
- Update `Environment.Status` with Ready, Pending, Error, and lists of created/failed resources
- Add sample CRs and test the full flow with `kubectl apply`

By the end, you will understand how parent-child relationships work in Kubernetes, how ownership enables garbage collection, and how finalizers allow cleanup before resource deletion.

---

## Prerequisites

Before starting, ensure you have completed **Phase 1**:

- Environment and KafkaTopic APIs exist
- KafkaTopic controller has minimal reconcile logic
- `make install` and `make run` work
- You can create a KafkaTopic and see the controller react

---

## Todo List

Use this checklist as you work through the phase. Check off each item when done.

- [ ] **2.1** Scaffold PostgreSQL API with `kubebuilder create api`
- [ ] **2.2** Edit PostgresqlDatabase types: add spec (name, connection refs) and status
- [ ] **2.3** Scaffold MongoDB API with `kubebuilder create api`
- [ ] **2.4** Edit MongodbDatabase types: add spec (name, connection refs) and status
- [ ] **2.5** Run `make generate` and `make manifests`
- [ ] **2.6** Update Environment controller RBAC for child resources (KafkaTopic, PostgresqlDatabase, MongodbDatabase)
- [ ] **2.7** Implement reconciliation: list desired children from Environment.Spec, create/update KafkaTopic CRs
- [ ] **2.8** Implement reconciliation: create/update PostgresqlDatabase and MongodbDatabase CRs
- [ ] **2.9** Add `SetControllerReference` so children are owned by the Environment
- [ ] **2.10** Add `Owns()` to the controller so it watches child CRs and re-reconciles when they change
- [ ] **2.11** Update Environment.Status (Ready, Pending, Error, CreatedResources, FailedResources)
- [ ] **2.12** Add finalizer for cascade deletion (optional: clean up external resources before removal)
- [ ] **2.13** Add sample Environment CR and child samples in `config/samples/`
- [ ] **2.14** Run `make install`, `make run`, apply samples, and verify child CRs are created
- [ ] **2.15** Test deletion: delete Environment and verify children are removed (garbage collection)

---

## Step-by-Step Explanations

### 2.1 Scaffold PostgreSQL API

**What to do**: Run `kubebuilder create api` for the PostgreSQL resource. Use group `postgresql`, version `v1alpha1`, kind `PostgresqlDatabase`. Accept both resource and controller.

```bash
kubebuilder create api --group postgresql --version v1alpha1 --kind PostgresqlDatabase --resource --controller
```

**Why it matters**: The Environment spec references PostgreSQL databases. Each database will be a separate CR that the PostgreSQL controller (Phase 3) will provision. The Environment controller creates these CRs.

**What to learn**: Each technology gets its own API group: `postgresql.shippercd.io`, `mongodb.shippercd.io`, `kafka.shippercd.io`. This keeps types isolated and RBAC scoped.

---

### 2.2 Edit PostgresqlDatabase Types

**What to do**: Open `api/postgresql/v1alpha1/postgresqldatabase_types.go`. Add to the Spec: `Name` (string), and optionally connection details (e.g. `ConnectionSecretRef` pointing to a Secret with host, port, credentials). Add to the Status: `Ready`, `Message`, `Error`, and optionally `Conditions` using `metav1.Condition`.

**Why it matters**: The Spec defines what the user (or Environment controller) wants. The Status is updated by the PostgreSQL controller in Phase 3. For Phase 2, the Environment controller only creates the CR; the Status can stay empty.

**What to learn**: Keep the Spec minimal for now. Connection details can be added in Phase 3 when provisioning logic is implemented.

---

### 2.3 Scaffold MongoDB API

**What to do**: Run `kubebuilder create api` for the MongoDB resource. Use group `mongodb`, version `v1alpha1`, kind `MongodbDatabase`. Accept both resource and controller.

```bash
kubebuilder create api --group mongodb --version v1alpha1 --kind MongodbDatabase --resource --controller
```

**Why it matters**: Same pattern as PostgreSQL. The Environment declares MongoDB databases; the Environment controller creates MongodbDatabase CRs; the MongoDB controller (Phase 3) will provision them.

**What to learn**: Consistent API design across technologies simplifies the Environment controller logic.

---

### 2.4 Edit MongodbDatabase Types

**What to do**: Edit `api/mongodb/v1alpha1/mongodbdatabase_types.go`. Add Spec fields: `Name`, and optionally `ConnectionSecretRef`. Add Status: `Ready`, `Message`, `Error`, `Conditions`.

**Why it matters**: Mirrors the PostgresqlDatabase structure. The Environment controller creates both with similar patterns.

**What to learn**: Reuse the same status shape (Ready, Message, Error, Conditions) across technology CRDs for consistency.

---

### 2.5 Run make generate and make manifests

**What to do**: After adding PostgreSQL and MongoDB APIs, run `make generate` and `make manifests`. Update `cmd/main.go` to register the new schemes and controllers (Kubebuilder may add scaffold markers). Run `make test` to verify.

**Why it matters**: New CRDs must be generated. The manager must register the new types and controllers.

**What to learn**: Every new API adds scheme registrations and controller setup. The default `make run` runs all controllers together for local development.

---

### 2.6 Update Environment Controller RBAC

**What to do**: Add RBAC markers to the Environment controller so it can create, update, patch, and delete KafkaTopic, PostgresqlDatabase, and MongodbDatabase resources. Example:

```go
// +kubebuilder:rbac:groups=kafka.shippercd.io,resources=kafkatopics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kafka.shippercd.io,resources=kafkatopics/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=postgresql.shippercd.io,resources=postgresqldatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgresql.shippercd.io,resources=postgresqldatabases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mongodb.shippercd.io,resources=mongodbdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mongodb.shippercd.io,resources=mongodbdatabases/status,verbs=get;update;patch
```

Then run `make manifests` to regenerate `config/rbac/role.yaml`.

**Why it matters**: The Environment controller acts on child resources. Without RBAC, `client.Create()` will fail with "forbidden".

**What to learn**: RBAC is declarative — you declare what the controller needs; the generated Role/ClusterRole grants it.

---

### 2.7 Implement Reconciliation: Create KafkaTopic CRs

**What to do**: In the Environment controller's `Reconcile` function:

1. Fetch the Environment object.
2. Loop over `env.Spec.KafkaTopics`.
3. For each KafkaTopic in the spec, build a `kafkav1alpha1.KafkaTopic` CR with:
   - `ObjectMeta`: `Name` (derive from Environment name + topic name to ensure uniqueness), `Namespace` (same as Environment).
   - `Spec`: `Name`, `Partitions`, `ReplicationFactor` from the spec item.
4. Use `r.Get()` to check if the CR already exists.
5. If not, use `r.Create()` to create it. If it exists and the spec changed, use `r.Update()` or `r.Patch()`.

**Why it matters**: This is the core of the Environment controller — translating Environment.Spec into child CRs.

**What to learn**: Naming strategy matters. Child CRs are typically named after the parent and the item (e.g. `{env-name}-{topic-name}`) to avoid collisions. Use `client.ObjectKeyFromObject` and consistent naming.

---

### 2.8 Implement Reconciliation: Create PostgresqlDatabase and MongodbDatabase CRs

**What to do**: Repeat the pattern for `env.Spec.PostgreSQLDatabases` and `env.Spec.MongoDBDatabases`. Create `PostgresqlDatabase` and `MongodbDatabase` CRs with the appropriate spec from each item. Use the same namespace as the Environment.

**Why it matters**: The Environment is the single source of truth. One Environment CR declares all databases and topics; the controller creates all child CRs.

**What to learn**: The reconciliation logic is similar for each technology. You can extract a helper or use a generic pattern to reduce duplication.

---

### 2.9 Add SetControllerReference

**What to do**: Before creating each child CR, call `ctrl.SetControllerReference(&env, child, r.Scheme)`. This sets `child.OwnerReferences` to point to the Environment. Kubernetes will then automatically delete the child when the Environment is deleted (garbage collection).

**Why it matters**: Without ownership, deleting an Environment would leave orphaned KafkaTopic, PostgresqlDatabase, and MongodbDatabase CRs. With ownership, Kubernetes garbage-collects them.

**What to learn**: `SetControllerReference` requires the controller to have permission to set `ownerReferences` on the child. The child must be in the same namespace as the owner (or the owner must be cluster-scoped). The scheme is used to set the `apiVersion` and `kind` of the owner.

---

### 2.10 Add Owns() to the Controller

**What to do**: In `SetupWithManager`, add `Owns()` for each child type so the controller re-reconciles the Environment when any child changes:

```go
return ctrl.NewControllerManagedBy(mgr).
    For(&corev1alpha1.Environment{}).
    Owns(&kafkav1alpha1.KafkaTopic{}).
    Owns(&postgresqlv1alpha1.PostgresqlDatabase{}).
    Owns(&mongodbv1alpha1.MongodbDatabase{}).
    Named("environment").
    Complete(r)
```

**Why it matters**: When a KafkaTopic's status changes (e.g. becomes Ready), the Environment controller should re-run to update `Environment.Status`. `Owns()` creates a watch on child resources and enqueues the parent for reconciliation when children change.

**What to learn**: `For()` watches the primary resource. `Owns()` watches owned resources and maps events back to the owner. This keeps the Environment status in sync with child state.

---

### 2.11 Update Environment.Status

**What to do**: After creating/updating children, compute the Environment status:

- **Ready**: `true` if all children are Ready (or you can use a condition).
- **PendingResources**: names of children that exist but are not yet Ready.
- **CreatedResources**: names of children that were created or updated.
- **FailedResources**: names of children that failed (e.g. Create returned an error).
- **Message** / **Error**: human-readable summary.

Use `r.Status().Update(ctx, &env)` to persist. Re-fetch the Environment before updating to avoid conflicts.

**Why it matters**: Users and operators need visibility. `kubectl get environment` and `kubectl describe environment` should show whether children were created and their state.

**What to learn**: Status updates can conflict if another controller updates the same resource. Use `Patch` with strategic merge or re-fetch before update. Consider `metav1.Condition` for structured status (see Phase 1 and AGENTS.md).

---

### 2.12 Add Finalizer (Optional)

**What to do**: Add a finalizer (e.g. `environments.core.shippercd.io/finalizer`) to the Environment. When the Environment is marked for deletion, `DeletionTimestamp` is set. In Reconcile, check if the object is being deleted; if so, ensure all children are removed (or clean up external resources), then remove the finalizer so Kubernetes can complete the deletion.

**Why it matters**: Sometimes you need to run logic *before* the resource is deleted (e.g. delete external resources, wait for child deletion). A finalizer blocks deletion until you remove it.

**What to learn**: With `SetControllerReference`, children are garbage-collected automatically. A finalizer is only needed if you have additional cleanup (e.g. external APIs, or if you want to control the order of deletion explicitly).

---

### 2.13 Add Sample CRs

**What to do**: Create or update `config/samples/core_v1alpha1_environment.yaml` with an Environment that has KafkaTopics, PostgreSQLDatabases, and MongoDBDatabases in its spec. Ensure the YAML matches your types (e.g. `postgresqlDatabases`, `mongodbDatabases`, `kafkaTopics`).

**Why it matters**: Sample CRs document usage and make it easy to test. `kubectl apply -k config/samples/` should apply everything.

**What to learn**: Keep samples minimal but valid. Users will copy them as starting points.

---

### 2.14 Run and Verify

**What to do**: Run `make install` (or `kubectl apply -f config/crd/bases/`), then `make run`. In another terminal, apply the sample Environment. Watch the controller logs. List child CRs: `kubectl get kafkatopics`, `kubectl get postgresqldatabases`, `kubectl get mongodbdatabases`. Verify they exist and have the correct owner reference.

**Why it matters**: End-to-end verification. The controller creates children; you see them in the cluster.

**What to learn**: Use `kubectl get environment -o yaml` to see status. Use `kubectl get kafkatopic -o yaml` to see `metadata.ownerReferences` pointing to the Environment.

---

### 2.15 Test Deletion

**What to do**: Delete the Environment: `kubectl delete environment <name>`. Verify that all child KafkaTopics, PostgresqlDatabases, and MongodbDatabases are also deleted (garbage collection).

**Why it matters**: Ownership ensures no orphans. Deleting the parent should remove the children.

**What to learn**: Kubernetes garbage collector runs asynchronously. Children may disappear a moment after the parent. If they don't, check that `SetControllerReference` was called and that `ownerReferences` are set on the children.

---

## Concepts to Internalize

| Concept | Description |
|--------|-------------|
| **Parent-child (owner-owned)** | A parent resource "owns" child resources via `metadata.ownerReferences`. Kubernetes garbage-collects children when the parent is deleted. |
| **SetControllerReference** | Sets the owner reference on a child so the parent is the "controller" owner. Required for `Owns()` and garbage collection. |
| **Owns()** | Tells the controller to watch owned resources. When a child changes, the parent is enqueued for reconciliation. |
| **Finalizer** | A string in `metadata.finalizers` that blocks deletion. The controller removes it when cleanup is done; only then does the resource get deleted. |
| **Idempotent reconciliation** | Reconcile should be safe to run multiple times. Check current state before Create/Update; don't assume the world is empty. |

---

## Troubleshooting

- **"forbidden" when creating child CR**: Add RBAC markers for the child resource's API group and run `make manifests`. Ensure the controller's ServiceAccount has the ClusterRole applied.
- **Children not deleted when Environment is deleted**: Verify `SetControllerReference` is called before `Create`. Check `kubectl get kafkatopic -o yaml` for `ownerReferences`.
- **Reconcile not triggered when child changes**: Add `Owns(&kafkav1alpha1.KafkaTopic{})` (and similar) in `SetupWithManager`. Without `Owns()`, the controller only reconciles when the Environment changes.
- **Update conflict on Environment.Status**: Re-fetch the Environment before updating status. Use `Patch` instead of `Update` if needed. Ensure you're not overwriting concurrent updates.
- **Child CR name collision**: Use a naming scheme that includes both the Environment name and the item name (e.g. `{env}-{topic}`). Ensure uniqueness within the namespace.

---

## Next Steps

After completing Phase 2, you are ready for Phase 3: implementing provisioning logic in each technology controller. The Kafka controller will use Sarama to create topics; the PostgreSQL controller will use pgx to create databases; the MongoDB controller will use the mongo-go-driver to create databases and users.
