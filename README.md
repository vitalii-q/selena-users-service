🧩 Selena Users Service

<!--![Docker](https://img.shields.io/badge/Docker-Containers-blue)-->
![CI/CD](https://img.shields.io/badge/CI/CD-GitHub_actions-c01c1c)

---

📌 Overview

users-service is a **cloud-oriented microservice**, responsible for managing user data and authentication-related logic.

The service runs in AWS cloud environment with:

- horizontal scalability
- secure secret management
- isolated infrastructure
- container-based deployment

<!--This repository contains both:

- application code (Go)
- infrastructure integration logic (Docker, CI/CD, migrations)-->

---

🚀 Key Characteristics

- Stateless service (horizontal scaling via ASG)
- Runs on EC2 inside private subnets
- Docker-based deployment via Amazon ECR
- Integrated with AWS Secrets Manager
- Uses AWS RDS (PostgreSQL) as primary database
- CI/CD pipeline via GitHub Actions


---

🏗️ How It Runs in AWS

                                Internet
                                    │
                            ┌───────▼────────┐
                            │   Public ALB   │
                            └───────┬────────┘
                                    │
        ┌───────────────────────────▼────────────────────────────────┐
        │                     Auto Scaling Group                     │
        │                                                            │
        │   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   │
        │   │     EC2      │   │     EC2      │   │     EC2      │   │
        │   │  ┌────────┐  │   │  ┌────────┐  │   │  ┌────────┐  │   │
        │   │  │ Docker │  │   │  │ Docker │  │   │  │ Docker │  │   │
        │   │  │ ┌────┐ │  │   │  │ ┌────┐ │  │   │  │ ┌────┐ │  │   │
        │   │  │ │MSVC│ │  │   │  │ │MSVC│ │  │   │  │ │MSVC│ │  │   │
        │   │  │ └────┘ │  │   │  │ └────┘ │  │   │  │ └────┘ │  │   │
        │   │  └────────┘  │   │  └────────┘  │   │  └────────┘  │   │
        │   └──────────────┘   └──────────────┘   └──────────────┘   │
        └───────────────────────────┬───────────────────────┬────────┘
                                    │                       ▲
                                    │                       └───────────┐
                                    ▼                                   ▼ 
                        ┌──────────────────────┐             ┌──────────────────────┐
                        │     NAT Instance     │             │   RDS (PostgreSQL)   │ 
                        └───────────┬──────────┘             └──────────────────────┘
                                    │
                                    ▼
                                 Internet

<!--
---

🔄 Request Flow

    Client
      │
      ▼
    Public ALB
      │
      ▼
    Auto Scaling Group (Available EC2)
      │
      ▼
    users-service (Docker)
      │
      ├──────────────► RDS (PostgreSQL)
      │
      └──────────────► Internal ALB ───► hotels-service
-->

---

☁️ Cloud Integration (AWS)

💻 Compute

- Runs on EC2 instances managed by Auto Scaling Group
- Scaling:
    - min: 1
    - max: 3

Each instance:

- is based on a custom AMI (Packer)
- runs Docker container with the service


📦 Containerization

- Docker image is built via CI
- Stored in Amazon ECR
- Pulled during EC2 startup


Dockerfile
Located at:

users-service/Dockerfile


🔐 Secrets Management

Managed via **AWS Secrets Manager**

Contains:
- DB credentials
- environment variables
- service configs

Secrets are injected into the container at runtime


🗄️ Database

- AWS RDS (PostgreSQL)
- Runs in private subnet
- Accessible only from users-service

Migrations

users-service/db/migrate.sh

Rollback

users-service/db/rollback.sh


⚖️ Load Balancing

Public ALB
Routes traffic:

users-service.selena-aws.com

Health Check
- Endpoint: /health
- Used by ALB to determine instance health


🌐 Internal Communication

Service can be accessed internally via:

users.internal.selena

Used for communication with other services (e.g. hotels-service)

---

🧱 Project Structure

    users-service/
    │── .github/workflows/     # CI/CD pipeline
    │── _docker/               # container entrypoint
    │── cmd/                   # CLI commands (seed, clean)
    │── db/                    # migrations & scripts
    │── internal/
    │   ├── bootstrap/         # app initialization
    │   ├── config/            # configuration loading
    │   ├── database/          # DB connection
    │   ├── handlers/          # HTTP handlers
    │   ├── services/          # business logic
    │   ├── router/            # routes
    │   └── server/            # HTTP server
    │
    │── tests/                 # integration tests
    │── main.go
    │── Dockerfile

---

🧪 Local Development

Run with Docker Compose

docker-compose up --build

Run migrations

cd users-service/db
./migrate.sh

Seed database

go run cmd/seed/main.go

---

⚙️ Configuration

Configuration is loaded from:
- environment variables
- AWS Secrets Manager (in cloud)

Example:

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret

---

⚠️ Notes

- Service is stateless → safe to scale horizontally
- No direct public access to EC2 instances
- All traffic goes through ALB
- Database is not publicly accessible

<!--
---

📈 Future Improvements

- Add distributed tracing
- Add retries & circuit breakers
- Introduce service mesh (future EKS migration)
- Improve CI/CD with OIDC-->
