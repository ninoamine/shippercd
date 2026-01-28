# Shipper Controller - v1alpha1 API Development Guide

## Overview
The `api/shipper-controller/v1alpha1` package defines the Custom Resource Definitions (CRDs) and API types for the shipper controller. This guide covers the API structure, resource types, and development patterns.

## API Execution Flow

```
API Request → Validation → Admission Webhooks → Reconciliation → Status Update
```

## Key Libraries & Dependencies

### Core Kubernetes Libraries
- **`sigs.k8s.io/controller-runtime/pkg/client`** - Kubernetes client for resource operations
  - Type-safe client for custom resources
  - Supports CRUD operations on CRD types

- **`k8s.io/apimachinery`** - Kubernetes API machinery
  - Meta types for resource versioning
  - Field selectors and label handling

### CRD & Schema
- **`sigs.k8s.io/controller-runtime/pkg/scheme`** - Resource scheme registration
  - Register custom resource types
  - Enable serialization/deserialization

- **`sigs.k8s.io/controller-tools`** - CRD generation from Go types
  - Generates YAML manifests from struct tags
  - Validates CRD schema definitions

### Validation & Webhooks
- **`sigs.k8s.io/controller-runtime/pkg/webhook`** - Admission webhooks
  - Validate resources before persistence
  - Mutate resources on creation/update

### Status Management
- **`k8s.io/api/core/v1`** (stdlib patterns) - Status subresources
  - Separate status from spec
  - Track conditions and phase

## Key Responsibilities

### 1. **Resource Type Definitions**
Define CRD types with:
- `Spec` - Desired state (user-provided)
- `Status` - Current state (controller-managed)
- Metadata - Labels, annotations, ownership

### 2. **Field Validation**
Implement validation rules:
- Required field checks
- Field format validation (regex, ranges)
- Cross-field validation rules
- CRD validation markers

### 3. **Admission Webhooks**
Setup webhook handlers for:
- ValidatingWebhook - Reject invalid resources
- MutatingWebhook - Set defaults and normalize values

### 4. **Status Subresource**
Manage resource status:
- Conditions (Ready, Error, Progressing)
- Phase tracking (Pending, Active, Completed)
- Observed generation for optimistic concurrency

### 5. **Schema Documentation**
Document all fields with:
- Purpose and usage
- Valid values and constraints
- Examples and references

## v1alpha1 Resource Lifecycle

```
Create Request
    ↓
Validation Webhook
    ↓
Mutating Webhook (defaults)
    ↓
Persist to etcd
    ↓
Controller Reconciliation
    ↓
Status Update
    ↓
Ready for Use
```

## Development Checklist
- [ ] All resource types are defined with proper struct tags
- [ ] Validation rules are comprehensive and documented
- [ ] Webhooks handle edge cases and provide clear error messages
- [ ] Status conditions follow standard Kubernetes patterns
- [ ] CRD manifests are generated and up-to-date
- [ ] API documentation is complete with examples
