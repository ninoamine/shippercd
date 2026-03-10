# Phase 1: Foundations with Kubebuilder

**Estimated duration**: 1–2 weeks  
**Learning goal**: Kubebuilder workflow, controller-runtime, watches, reconcile loop.

---

## Overview

Phase 1 sets up the core structure of your ShipperCD project. You will:

- Initialize a Kubebuilder project with a custom domain
- Create your first two Custom Resource Definitions (CRDs): Environment and KafkaTopic
- Implement a minimal controller that reacts to KafkaTopic resources
- Run everything locally against a Kubernetes cluster

By the end, you will understand how controllers watch resources, how the reconcile loop works, and how CRDs extend the Kubernetes API.

---

## Prerequisites

Before starting, ensure you have:

- **Go** installed (1.21+)
- **Kubebuilder** installed — follow the [official installation guide](https://book.kubebuilder.io/quick-start.html#installation)
- **kubectl** configured to talk to a cluster (e.g. kind, minikube, or a cloud cluster)
- **Docker** (if using kind or minikube)

---

## Todo List

Use this checklist as you work through the phase. Check off each item when done.

- [ ] **1.1** Install Kubebuilder and verify it works
- [ ] **1.2** Create a new directory for the project and `cd` into it
- [ ] **1.3** Run `kubebuilder init` to scaffold the project
- [ ] **1.4** Explore the generated structure (go.mod, main.go, Makefile, config/)
- [ ] **1.5** Create the Environment API with `kubebuilder create api`
- [ ] **1.6** Edit Environment types: add spec (databases, topics, limits) and status
- [ ] **1.7** Run `make generate` and `make manifests`
- [ ] **1.8** Create the KafkaTopic API with `kubebuilder create api`
- [ ] **1.9** Run `make generate` and `make manifests` again
- [ ] **1.10** Implement minimal reconcile logic in the KafkaTopic controller
- [ ] **1.11** Run `make install` to install CRDs into your cluster
- [ ] **1.12** Run `make run` to start the controller
- [ ] **1.13** Create a sample KafkaTopic resource and apply it with kubectl
- [ ] **1.14** Observe the controller logs and verify it reconciles

---

## Step-by-Step Explanations

### 1.1 Install Kubebuilder

**What to do**: Install Kubebuilder using the official instructions for your OS.

**Why it matters**: Kubebuilder is the framework that generates the boilerplate for CRDs and controllers. It uses controller-runtime under the hood and follows Kubernetes conventions.

**What to learn**: Kubebuilder is one of several options (Operator SDK, KUDO, etc.) for building operators. It is widely used and well-documented.

---

### 1.2 Create Project Directory

**What to do**: Create a new empty directory (e.g. `shippercd`) and `cd` into it. Do not run `go mod init` yourself — Kubebuilder will do that.

**Why it matters**: Kubebuilder expects to run in an empty or new directory. It will create `go.mod`, `main.go`, and the project layout.

---

### 1.3 Run kubebuilder init

**What to do**: Run the init command with your domain. The domain becomes part of the API group (e.g. `core.envcd.io`).

**Why it matters**: This scaffolds the entire project: `main.go` with manager setup, `Makefile` with targets for generate/manifests/install/run, `config/` with RBAC and deployment manifests, and `PROJECT` metadata.

**What to learn**: The domain is used for API groups. All your CRDs will live under `*.envcd.io`. The manager is the process that runs your controllers and talks to the Kubernetes API.

---

### 1.4 Explore the Generated Structure

**What to do**: Open and skim the generated files. Focus on:

- `cmd/main.go` — how the manager is created and started
- `Makefile` — targets like `generate`, `manifests`, `install`, `run`
- `config/` — CRD bases, RBAC, default manager deployment

**Why it matters**: Understanding the layout helps you know where to add code and what each part does. You will extend these files, not replace them.

**What to learn**: Kubebuilder uses markers (comments) in Go files to know where to inject code. Do not remove `+kubebuilder:scaffold` markers.

📖 **Beginner-friendly deep dive**: See [Phase 1, Step 1.4: Exploring the Kubebuilder Structure](step1.4-exploring-kubebuilder-structure.md) for a detailed walkthrough of the generated structure and the Go concepts used (packages, imports, flags, scheme, manager, scaffold markers).

---

### 1.5 Create the Environment API

**What to do**: Run the create-api command for the Environment resource. Use the flags for group, version, kind, and ensure you request both resource and controller.

**Why it matters**: This generates the Environment CRD types and a controller skeleton. The Environment will be your main entity — users declare databases and topics in it.

**What to learn**: In Kubernetes, an API group + version + kind uniquely identify a resource type. Your Environment will be `Environment` in group `core.shippercd.io`, version `v1alpha1`.

📖 **Beginner-friendly guide**: See [Phase 1, Step 1.5: Creating the Environment API](step1.5-creating-the-environment-api.md) for concepts (API group, version, kind), what the command does, what gets generated, and a checklist so you can do it yourself.

---

### 1.6 Edit Environment Types

**What to do**: Open the Environment types file. Add fields to the Spec and Status structs:

- **Spec**: Think about what an environment contains — a list of databases (PostgreSQL, MongoDB), a list of topics (Kafka), and optional limits (e.g. max databases per env).
- **Status**: Reflect the state — Ready, Pending, Error, and perhaps a message or list of created resources.

**Why it matters**: The Spec is what users define. The Status is what the controller updates to reflect reality. This is the desired state vs. actual state pattern.

**What to learn**: CRD types use struct tags for serialization (JSON) and for OpenAPI validation. Kubebuilder adds markers for CRD generation (e.g. `+kubebuilder:subresource:status`).

📖 **Beginner-friendly guide**: See [Phase 1, Step 1.6: Editing Environment Types](step1.6-editing-environment-types.md) for Spec/Status design, struct tags, markers, and verification checklist.

---

### 1.7 Run make generate and make manifests

**What to do**: After editing types, run these two Make targets. Fix any errors (e.g. missing imports, invalid markers).

**Why it matters**: `make generate` runs the controller-gen tool to produce DeepCopy methods and other generated code. `make manifests` produces the CRD YAML files from your types.

**What to learn**: Kubebuilder relies on code generation. You edit the source types; the tooling generates the rest. Always run these after changing types.

---

### 1.8 Create the KafkaTopic API

**What to do**: Run the create-api command for KafkaTopic. Use a different group (e.g. kafka) so it lives under `kafka.envcd.io`.

**Why it matters**: KafkaTopic represents a single Kafka topic. Later, the Environment controller will create KafkaTopic CRs; the Kafka controller will watch them and provision topics in Kafka.

**What to learn**: You can have multiple API groups in one project. Each gets its own directory under `api/`.

---

### 1.9 Run make generate and make manifests Again

**What to do**: Regenerate after adding KafkaTopic. Check that new CRD YAML files appear in `config/crd/bases/`.

**Why it matters**: Each new API adds new types and a new controller. The Makefile compiles everything together.

**What to learn**: The manager registers all controllers. When you run `make run`, both the Environment and KafkaTopic controllers will be active (even if their logic is minimal).

---

### 1.10 Implement Minimal Reconcile Logic

**What to do**: Open the KafkaTopic controller. Find the Reconcile function. Implement logic that:

- Logs when a KafkaTopic is reconciled (with its name and namespace)
- Optionally updates the Status (e.g. set a condition or message)

Do not call Kafka yet — just make the controller react and update status.

**Why it matters**: The reconcile loop is the heart of a controller. It is called whenever a watched resource changes. Your job is to make the cluster state match the desired state (the Spec).

**What to learn**: Reconcile receives a request (namespace + name). You fetch the object, check its state, and optionally update it or create/update other resources. Returning an error triggers a retry. Returning success with `Requeue: true` schedules another reconcile after a delay.

---

### 1.11 Run make install

**What to do**: Run the install target. This applies the CRD manifests to your cluster.

**Why it matters**: CRDs must be installed before you can create resources of that type. Without them, `kubectl apply` of an Environment or KafkaTopic will fail.

**What to learn**: CRDs extend the Kubernetes API. Once installed, you can create custom resources just like Pods or Services.

---

### 1.12 Run make run

**What to do**: Start the controller. It will run in the foreground and connect to your cluster (using your kubeconfig).

**Why it matters**: The controller needs to be running to react to resources. In production, it runs as a Deployment; locally, `make run` is convenient for development.

**What to learn**: The controller uses a client that watches the API server. When you create or update a KafkaTopic, the watch fires and the reconcile loop is invoked.

---

### 1.13 Create a Sample KafkaTopic

**What to do**: Create a YAML file for a KafkaTopic (use the samples in `config/samples/` as a template). Apply it with `kubectl apply -f`.

**Why it matters**: This triggers your controller. You will see the reconcile loop run in the logs.

**What to learn**: Custom resources are just YAML like any other Kubernetes object. They have `apiVersion`, `kind`, `metadata`, and `spec`.

---

### 1.14 Observe Controller Logs

**What to do**: Watch the controller output. You should see log lines when the KafkaTopic is created and reconciled. If you updated status, run `kubectl get kafkatopic -o yaml` to see the status field.

**Why it matters**: This confirms the full loop works: watch → reconcile → (optionally) update.

**What to learn**: Controllers are event-driven. They do not poll; they react to changes. The controller-runtime handles the watch, deduplication, and queueing for you.

---

## Concepts to Internalize

| Concept | Description |
|--------|-------------|
| **CRD** | Custom Resource Definition — extends the Kubernetes API with your own resource types |
| **Controller** | A loop that watches resources and tries to make the cluster state match the desired state |
| **Reconcile** | The function that runs for each change; it receives the resource and can read/update/create other resources |
| **Manager** | The process that runs one or more controllers and shares the Kubernetes client and cache |
| **Spec vs Status** | Spec = desired state (user-defined). Status = actual state (controller-updated) |

---

## Troubleshooting

- **Controller does not start**: Check that your kubeconfig points to a valid cluster and that you have permissions to create/watch the CRDs.
- **Reconcile not called**: Ensure the CRD is installed and the resource has the correct `apiVersion` and `kind`.
- **make generate fails**: Check for syntax errors in your types and that all required markers are present.
- **CRD install fails**: Ensure no conflicting CRDs exist (e.g. from a previous install). You may need to delete old CRDs first.

---

## Next Steps

After completing Phase 1, you are ready for Phase 2: implementing the Environment controller to create KafkaTopic, PostgreSQL, and MongoDB CRs from an Environment spec, and adding finalizers for cascade deletion.
