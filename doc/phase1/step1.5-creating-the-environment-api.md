# Phase 1, Step 1.5: Creating the Environment API

**For**: Beginners in Go and Kubernetes  
**Goal**: Understand what it means to "create an API" with Kubebuilder, what you will run, and what gets generated — so you can do it yourself and learn.

---

## What You Are About to Do

In this step, you will run the Kubebuilder CLI to scaffold your first Custom Resource Definition (CRD) and its controller. The **Environment** is the main entity in ShipperCD: users will declare what databases and topics they want in an Environment, and your controllers will create them.

You will **not** edit any code yet. You will run a single command, answer a couple of prompts, and then explore what was generated. The actual type definitions (Spec, Status) come in Step 1.6.

---

## Concepts: API Group, Version, Kind

In Kubernetes, every resource type is identified by three things:

| Term | Meaning | Example |
|------|---------|---------|
| **API group** | A logical grouping of related resources. Combines with the project domain. | `core` → `core.shippercd.io` |
| **Version** | API version for compatibility (e.g. `v1`, `v1alpha1`, `v1beta1`). | `v1alpha1` (alpha = may change) |
| **Kind** | The type name, as written in YAML (PascalCase). | `Environment` |

Together they form the **API path** and the `apiVersion` in YAML. For example, an Environment resource will have:

- `apiVersion: core.shippercd.io/v1alpha1`
- `kind: Environment`

The domain (`shippercd.io`) comes from your `PROJECT` file. You chose it when you ran `kubebuilder init`. Kubebuilder combines `group` + `domain` to form the full group name.

---

## The Command You Will Run

Kubebuilder provides a subcommand to create a new API:

```
kubebuilder create api --group <group> --version <version> --kind <Kind>
```

### Understanding the Flags

- **`--group`**: The API group (without the domain). For Environment, use `core` — it will become `core.shippercd.io`.
- **`--version`**: The API version. Use `v1alpha1` for a new, experimental type. (`alpha` = likely to change; `beta` = more stable; `v1` = stable.)
- **`--kind`**: The resource type name in PascalCase. Use `Environment`.

### Optional Flags (Defaults Are Usually Fine)

- **`--resource`**: Whether to generate the CRD and types. Default: `true`. You want this.
- **`--controller`**: Whether to generate the controller. Default: `true`. You want this.

You can omit these; the defaults are what you need for a full CRD + controller.

### Interactive Prompts

When you run the command, Kubebuilder may ask:

- *Create Resource [y/n]* — Answer `y` (you want the CRD and Go types).
- *Create Controller [y/n]* — Answer `y` (you want the controller skeleton).

If you pass both `--resource` and `--controller` explicitly, it may skip the prompts.

---

## What Gets Generated

After you run the command, Kubebuilder will create and modify files. Here is what to expect — no code, just the layout and purpose.

### New Files and Folders

1. **`api/<group>/<version>/`** — A new directory for your API types.
   - `*_types.go` — Your Spec and Status structs (mostly empty stubs for now).
   - `groupversion_info.go` — Tells Kubebuilder the group and version.
   - `zz_generated.deepcopy.go` — **Do not edit.** Generated DeepCopy methods.

2. **`internal/controller/<kind>_controller.go`** — The controller skeleton.
   - A `Reconcile` function that runs when an Environment is created or updated.
   - Setup code that registers the controller with the manager.
   - Initially it does almost nothing; you will add logic in Step 1.10 (for KafkaTopic) and Phase 2 (for Environment).

3. **`config/crd/bases/`** — CRD YAML files.
   - One YAML file per CRD, describing the schema to Kubernetes.
   - Generated from your types and markers. You will regenerate these with `make manifests` after editing types.

4. **`config/samples/`** — Example Custom Resources.
   - A sample Environment YAML you can apply to test.
   - Minimal placeholder; you will improve it after defining the Spec.

### Modified Files

1. **`cmd/main.go`** — Kubebuilder injects code at the scaffold markers:
   - Registers your new types in the scheme.
   - Creates and registers the Environment controller with the manager.

2. **`config/rbac/role.yaml`** — New permissions for the controller to manage Environment resources.

3. **`PROJECT`** — Kubebuilder adds metadata about the new API.

**Important**: Do not remove the `+kubebuilder:scaffold` markers in `main.go`. Kubebuilder needs them to inject code when you add more APIs later.

---

## Go Concepts You Will See

When you open the generated files, you will encounter:

| Concept | Where You See It |
|---------|------------------|
| **Struct** | `Environment` has a `Spec` and `Status` struct. Structs group related fields. |
| **Embedding** | `EnvironmentSpec` and `EnvironmentStatus` may embed `metav1.TypeMeta` or similar. Embedding reuses another type's fields. |
| **JSON struct tags** | Fields have `` `json:"fieldName"` `` — how they map to/from YAML and JSON. |
| **Pointer receiver** | Methods like `func (r *EnvironmentReconciler) Reconcile(...)` use `*` so the method can modify the receiver. |
| **Context** | `Reconcile(ctx context.Context, req ctrl.Request)` — `context` carries deadlines and cancellation signals. |

You do not need to understand all of this before running the command. See [Go Concepts for Kubernetes Operators](../go-concepts/concepts.md) for detailed explanations. Use Step 1.6 and the Kubebuilder book to go deeper when you edit the types.

---

## Checklist: Do It Yourself

1. Open a terminal in your project root.
2. Run the `kubebuilder create api` command with:
   - `--group core`
   - `--version v1alpha1`
   - `--kind Environment`
3. Answer the prompts if Kubebuilder asks about resource and controller.
4. When the command finishes, explore:
   - `api/` — new folder structure and `*_types.go`
   - `internal/controller/` — the controller file
   - `config/crd/bases/` — the generated CRD YAML
   - `config/samples/` — the sample Environment
   - `cmd/main.go` — what was added at the scaffold markers
5. Run `make generate` and `make manifests` to ensure everything compiles and CRDs are up to date. Fix any errors if they appear.
6. Do **not** edit the types yet — that is Step 1.6.

---

## What Can Go Wrong?

- **"Failed to run pre-scaffold task"** — Check that you are in the project root (where `PROJECT` and `go.mod` live) and that `PROJECT` has a valid `domain` and `repo`.
- **Directory already exists** — You may have run the command before. Either use that API or remove the generated files and try again.
- **`make generate` or `make manifests` fails** — Often due to syntax or marker issues in the generated types. In this step the defaults should work; if not, read the error message and check the Kubebuilder book.

---

## Next Step

In **Step 1.6**, you will open the `*_types.go` file and add fields to the Environment **Spec** (what users define) and **Status** (what the controller updates). That is where you design what an Environment actually contains.
