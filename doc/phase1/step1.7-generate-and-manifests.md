# Phase 1, Step 1.7: Running make generate and make manifests

**For**: Beginners in Go and Kubernetes  
**Goal**: Understand what these Make targets do, what files they update, and why you must run them after changing types.

---

## What You Are Doing

After editing `*_types.go` in Step 1.6, the Go code and the generated artifacts can get out of sync. The **`make generate`** and **`make manifests`** targets regenerate everything from your types. You run them so that:

- DeepCopy methods match your structs
- CRD YAML reflects your Spec and Status
- Lint and tests pass

---

## make generate

### What It Does

Runs **controller-gen** in object mode to generate:

- **`api/<group>/<version>/zz_generated.deepcopy.go`** — `DeepCopy`, `DeepCopyInto`, and `DeepCopyObject` methods for all your types.

Kubernetes controllers need DeepCopy to avoid mutating cached objects when they create copies for updates. Controller-gen reads your types and their markers, then generates implementations.

### When to Run

- After adding, removing, or changing fields in any struct used in a CRD (Spec, Status, nested types).
- Whenever `make lint` or `make test` fails with errors in generated code.

### Command

```bash
make generate
```

---

## make manifests

### What It Does

Runs **controller-gen** to generate:

- **`config/crd/bases/*.yaml`** — CRD definitions with OpenAPI schema from your types and markers.
- **`config/rbac/role.yaml`** — RBAC permissions for the controllers (roles may be updated when you add APIs).
- Webhook manifests (if webhooks are enabled).

The CRD YAML is what `kubectl apply` uses when you run `make install`. It tells the Kubernetes API server the schema of your custom resources (required fields, types, validation rules).

### When to Run

- After any change to `*_types.go` that affects the CRD schema (Spec, Status, markers like `+optional`, `+kubebuilder:validation:*`).
- Before `make install` or `make deploy`.

### Command

```bash
make manifests
```

---

## Order and Combined Use

Run both, usually in this order:

1. `make generate` — Updates generated Go code.
2. `make manifests` — Updates CRD and RBAC YAML.
3. `make fmt` — Ensures formatting is correct.
4. `make lint` — Verifies everything passes.

The Makefile often chains these: `make build`, `make run`, and `make test` depend on `manifests` and `generate`. So running `make test` will run them automatically — but it helps to run them explicitly after editing types so you can fix errors immediately.

---

## What to Check After Running

| Output | Location | What to verify |
|--------|----------|----------------|
| DeepCopy | `api/v1alpha1/zz_generated.deepcopy.go` | Methods exist for all your types (Environment, EnvironmentSpec, EnvironmentStatus, PostgreSQLDatabase, MongoDBDatabase, KafkaTopic). |
| CRD | `config/crd/bases/core.shippercd.io_environments.yaml` | `spec.properties.spec.properties` contains your fields (name, postgresqlDatabases, mongodbDatabases, kafkaTopics) with correct types. |
| RBAC | `config/rbac/role.yaml` | Includes rules for `environments.core.shippercd.io` if your controller manages them. |

---

## Common Errors and Fixes

| Error | Cause | Fix |
|-------|-------|-----|
| `undefined: X` or `X has no field Y` | Types changed but DeepCopy not regenerated | Run `make generate`. |
| `invalid marker` or `unknown marker` | Typo in a Kubebuilder marker (e.g. `+required` instead of `+kubebuilder:validation:Required`) | Fix the marker in `*_types.go`, then run `make generate` and `make manifests`. |
| `make lint` fails (gofmt) | Struct fields not aligned | Run `make fmt` or `gofmt -w api/`. |
| CRD schema missing your new fields | `make manifests` not run, or types not properly exported | Run `make manifests`; ensure new types are in the same package and referenced from Spec/Status. |

---

## Checklist

- [ ] `make generate` runs without errors
- [ ] `make manifests` runs without errors
- [ ] `config/crd/bases/core.shippercd.io_environments.yaml` contains your Spec and Status fields
- [ ] `make fmt` and `make lint` pass
- [ ] `make test` passes (optional but recommended)

---

## Next Step

In **Step 1.8**, you will create the KafkaTopic API with `kubebuilder create api`. After that, you will run `make generate` and `make manifests` again (Step 1.9) so the new CRD and controller are included.
