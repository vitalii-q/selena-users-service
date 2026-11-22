# Users Service Kubernetes Configuration

This directory is the canonical location for Kubernetes manifests and the Helm chart owned by `users-service`.

The local Kubernetes application connects to the existing Selena Docker PostgreSQL through `host.minikube.internal:9265`. Do not create another PostgreSQL server for this service. The local Kubernetes Secret `users-db-secret` must contain credentials for the isolated `users_k8s` database and role; never commit or print Secret values.

The base application resources target namespace `selena-dev` and use a ClusterIP Service. The Deployment and chart default to two application replicas for the service-balancing exercise. PostgreSQL remains the existing Docker service and is not scaled by this configuration. A one-shot migration/seed Job runs before the Deployment so application Pod restarts do not rerun migrations.

The base manifests and chart do not create a database, PVC, Ingress, or cloud resource. The chart's image pull policy is set for the preloaded local Minikube image; a future registry workflow can supply different image values without changing application code.
