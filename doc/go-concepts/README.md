# Go Concepts

Standalone reference for Go language concepts used in the ShipperCD project.

- **[concepts.md](concepts.md)** — Package, import, struct, pointer, error handling, `init`, context, interfaces, and more.
- **[interfaces-runtime-explained.md](interfaces-runtime-explained.md)** — How interfaces work at runtime: why `r.Get()` runs real code even though the interface only defines signatures. Step-by-step with the KafkaTopicReconciler.
- **[controller-reconcile-concepts.md](controller-reconcile-concepts.md)** — Go concepts in the controller and Reconcile function: struct embedding, pointer receivers, `ctrl.Request`/`ctrl.Result`, method chaining, client usage, and more.
- **[test-concepts.md](test-concepts.md)** — Test concepts (Ginkgo, Gomega, envtest) and Go patterns used in controller tests: Describe/Context/It, BeforeEach/AfterEach, Expect matchers, closures, package-level variables, and more.

These docs are written for beginners in Go who are building Kubernetes operators with Kubebuilder. Use them alongside the Phase 1 guides.
