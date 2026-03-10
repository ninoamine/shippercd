# Phase 1, Step 1.6: Editing Environment Types

**For**: Beginners in Go and Kubernetes  
**Goal**: Understand how to define the Spec and Status of your CRD, what struct tags and markers do, and how this affects the generated CRD schema.

---

## Review: Corrections Applied

When reviewing `api/v1alpha1/environment_types.go`, these fixes were applied:

| Issue | Correction |
|-------|------------|
| **JSON tags** | `json: "name"` (with space) → `json:"name"` (no space). The space can break JSON marshalling. |
| **Required marker** | `// +required` → `// +kubebuilder:validation:Required`. Only the latter is recognized by controller-gen. |

Your design choices (Spec with PostgreSQL, MongoDB, KafkaTopic lists; Status with Ready, Message, Error, and resource lists) are valid for Phase 1. For later phases, consider migrating Status to `metav1.Condition` per Kubernetes conventions (see AGENTS.md).

---

## What You Are Doing

In Step 1.5, Kubebuilder generated empty `EnvironmentSpec` and `EnvironmentStatus` structs. In this step, you add the fields that define what an Environment contains (Spec) and what the controller reports back (Status).

The **Spec** is the desired state — what the user declares. The **Status** is the observed state — what the controller updates to reflect reality. This is the Kubernetes desired-state vs actual-state pattern.

---

## Spec: What an Environment Contains

From the project plan, an Environment declares:

- **Databases** — PostgreSQL and MongoDB databases to create
- **Topics** — Kafka topics to create
- **Optional limits** — e.g. max databases per environment (you can add later)

For each database or topic, you need at least a name. For Kafka topics, you typically need partition count and replication factor.

### Nested Types

Rather than inlining all fields in `EnvironmentSpec`, you define small structs and reference them:

- `PostgreSQLDatabase` — fields for one PostgreSQL database
- `MongoDBDatabase` — fields for one MongoDB database  
- `KafkaTopic` — fields for one Kafka topic (name, partitions, replication factor)

Then `EnvironmentSpec` holds slices of these types. A slice `[]T` in Go is a list of `T`.

**Note on naming**: The struct `KafkaTopic` here describes the *spec of a topic within an Environment*. In Step 1.8 you will create a separate `KafkaTopic` CRD (a different API group). To avoid confusion, you could name this struct `KafkaTopicSpec` or `TopicReference`. Both exist in the same project but in different packages.

---

## Status: What the Controller Reports

The Status reflects what the controller has observed or done. Common approaches:

1. **Simple fields** — `Ready` (bool), `Message` (string), lists of created/failed/pending resources. Good for learning and Phase 1.
2. **Conditions** — `metav1.Condition[]` as recommended in Kubernetes API conventions. Used by many operators for Ready, Degraded, etc.

For Phase 1, simple fields are fine. In later phases you can migrate to `metav1.Condition` (see AGENTS.md).

### Useful Status Fields

- **Ready** — Boolean: is the Environment fully provisioned?
- **Message** — Human-readable summary
- **Error** — Last error message if something failed
- **CreatedResources** / **PendingResources** / **FailedResources** — Lists of resource names (or identifiers) to show progress

---

## Struct Tags and Markers

### JSON Tags

Fields use struct tags for serialization. The format is `` `json:"fieldName"` ``:

- **No space** after the colon: `json:"name"` ✓ — `json: "name"` ✗
- **omitempty** — Omit the field when it is the zero value (empty string, nil, 0)
- **omitzero** — (Go 1.24+) Similar to omitempty; omits zero values. Useful for optional fields without pointers

Example: `Name string `json:"name"`` makes the field appear as `"name"` in JSON/YAML.

### Kubebuilder Markers

Comments above types or fields control CRD generation:

| Marker | Purpose |
|--------|---------|
| `+kubebuilder:object:root=true` | Marks the type as a root CRD (not just a nested struct) |
| `+kubebuilder:subresource:status` | Status is a subresource (updates don't touch spec) |
| `+kubebuilder:validation:Required` | Field is required in the CRD schema |
| `+optional` | Field is optional (common in Kubernetes types) |

The marker `// +kubebuilder:validation:Required` must be used for required fields; `// +required` alone is not a valid Kubebuilder marker.

---

## After Editing

1. Run `make generate` — Regenerates DeepCopy methods for your new types in `zz_generated.deepcopy.go`. This is required when you add or change struct fields; without it, the generated code will be out of date and `make lint` may fail.
2. Run `make manifests` — Regenerates CRD YAML from your types and markers.
3. Run `make fmt` (or `gofmt -w api/`) — Ensures struct fields are properly aligned; `make lint` checks formatting.
4. Fix any errors — Typos in tags, unknown markers, or missing imports.

---

## Checklist: Verify Your Types

- [ ] JSON tags use `json:"name"` (no space after colon)
- [ ] Nested types (`PostgreSQLDatabase`, `MongoDBDatabase`, `KafkaTopic`) have appropriate fields
- [ ] `EnvironmentSpec` has slices of these types
- [ ] `EnvironmentStatus` has fields to reflect state (Ready, Message, lists, etc.)
- [ ] `// +kubebuilder:validation:Required` on required fields, not `// +required`
- [ ] `make generate` and `make manifests` run without errors
- [ ] `make fmt` and `make lint` pass (generated code up to date, formatting correct)

---

## Optional: Limits and More Fields

The foundations doc suggests optional limits (e.g. max databases per environment). You can add a `MaxDatabases` field. Using a pointer (`*int32`) makes it optional — `nil` means "no limit". With Go 1.24+, you can also use a value type with the `omitzero` tag.

---

## Next Step

In **Step 1.7**, you run `make generate` and `make manifests` to regenerate code and CRDs, then fix any errors. If those already succeed, you can move on to Step 1.8 (create the KafkaTopic API).
