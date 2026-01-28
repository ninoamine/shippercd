# Shipper Controller - Manager Development Guide

## Overview
The `cmd/shipper-controller/main.go` file initializes and starts the shipper controller manager. This guide focuses on the core development logic and execution flow.

## Main Execution Flow

```
main() → Parse Config → Setup Manager → Start Manager → Handle Signals
```

## Key Libraries & Dependencies

### Core Kubernetes Libraries
- **`sigs.k8s.io/controller-runtime`** - Framework for building Kubernetes controllers
  - Handles manager lifecycle, reconciliation loops, and caching
  - Provides client for API server communication
  
- **`k8s.io/client-go`** - Official Kubernetes Go client
  - REST client for Kubernetes API
  - Used by controller-runtime for API interactions

### Configuration & Flags
- **`flag`** (stdlib) - Parses command-line arguments
  - Define and parse controller flags
  
- **`os`** (stdlib) - OS operations
  - Environment variable access
  - Signal handling

### Logging & Observability
- **`go.uber.org/zap`** (or similar) - Structured logging
  - High-performance logging for controller events
  - Configurable log levels
  
- **`github.com/prometheus/client_golang`** - Prometheus metrics
  - Expose controller metrics for monitoring
  - Track reconciliation counts, latencies

### Graceful Shutdown
- **`os/signal`** (stdlib) - Signal handling
  - Listen for SIGINT, SIGTERM
  - Trigger graceful shutdown

### Context & Concurrency
- **`context`** (stdlib) - Context management
  - Propagate cancellation signals
  - Handle timeouts for operations

## Key Responsibilities

### 1. **Configuration & Flags**
Parse command-line arguments for:
- Kubernetes configuration (`-kubeconfig`)
- Watch namespace(s)
- Leader election settings
- Logging verbosity
- Sync intervals and timeouts

### 2. **Manager Setup**
Initialize the manager with:
- Kubernetes client configuration
- Resource reconcilers
- Webhook servers (if needed)
- Cache setup for watched resources

### 3. **Controller Registration**
Add all controllers to the manager:
- Each controller watches specific resources
- Controllers reconcile state when changes occur
- Multiple controllers can run concurrently

### 4. **Health & Readiness Probes**
Setup endpoints for:
- Liveness checks (is the manager running?)
- Readiness checks (is the manager ready to handle requests?)

### 5. **Signal Handling**
Gracefully shutdown when receiving:
- SIGINT (Ctrl+C)
- SIGTERM (termination signal)

## Manager Lifecycle

```
Start Manager
    ↓
Leader Election (if enabled)
    ↓
Sync Caches with API Server
    ↓
Start All Controllers
    ↓
Wait for Shutdown Signal
    ↓
Stop Controllers
    ↓
Exit
```

## Development Checklist
- [ ] All required flags are documented
- [ ] Manager metrics are exposed for monitoring
- [ ] Controllers handle errors and retries
- [ ] Graceful shutdown completes within timeout
- [ ] Leader election works in multi-replica deployments
