# Users Service Kubernetes Configuration

This directory is the canonical location for Kubernetes manifests and the Helm chart owned by `users-service`.

The local Kubernetes application connects to the existing Selena Docker PostgreSQL through `host.minikube.internal:9265`. Do not create another PostgreSQL server for this service. The local Kubernetes Secret `users-db-secret` must contain credentials for the isolated `users_k8s` database and role; never commit or print Secret values.

The Compose `.env` and Kubernetes Pod use different network paths to reach that PostgreSQL container. A Compose container on the Docker application network uses the service DNS name `users-db` and the container port `5432`. A Pod in Minikube cannot use that Compose-only DNS name; it connects through the Mac host alias `host.minikube.internal` and Docker's published host port `9265`, which forwards to PostgreSQL's container port `5432`. The application retains the existing variable name `USERS_POSTGRES_DB_PORT_INNER`, but its value is network-context-specific: `9265` in this Kubernetes ConfigMap. Do not copy the Compose host/port into the Kubernetes configuration.

Kubernetes intentionally uses the separate logical database `users_k8s`, not the Compose database `users_db`, so Kubernetes migration and seed operations do not touch the existing Compose application's database.

The base application resources target namespace `selena-dev` and use a ClusterIP Service. The Deployment and chart default to two application replicas for the service-balancing exercise. PostgreSQL remains the existing Docker service and is not scaled by this configuration. A one-shot migration/seed Job runs before the Deployment so application Pod restarts do not rerun migrations.

The current image returns HTTP 200 from `/test`; startup and liveness probes use that endpoint. Readiness runs a SQL `SELECT 1` with `psql` and environment variables sourced from the Secret. It does not print credentials.

The base manifests and chart do not create a database, PVC, Ingress, or cloud resource. The chart's image pull policy is set for the preloaded local Minikube image; a future registry workflow can supply different image values without changing application code.
