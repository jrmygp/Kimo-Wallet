# Kimo Wallet — System Architecture

## 1. Architecture Goals

The architecture is designed to demonstrate:

- Next.js + TypeScript frontend engineering.
- Mobile-first PWA architecture.
- Go microservices.
- gRPC for internal service-to-service communication.
- Protocol Buffers as the source of service contracts.
- Event-driven asynchronous processing.
- Reliable financial transaction handling.
- Independent service ownership and deployment.
- Observability and operational readiness.

The system intentionally starts small and can evolve without requiring a large infrastructure footprint.

## 2. High-Level Architecture

```text
                         ┌─────────────────────┐
                         │    Kimo Wallet PWA  │
                         │ Next.js + TypeScript │
                         └──────────┬──────────┘
                                    │
                              HTTPS / JSON
                                    │
                                    ▼
                         ┌─────────────────────┐
                         │    API Gateway      │
                         │         Go          │
                         └──────────┬──────────┘
                                    │
                              gRPC / Protobuf
                                    │
             ┌──────────────────────┼──────────────────────┐
             │                      │                      │
             ▼                      ▼                      ▼
      ┌──────────────┐      ┌──────────────┐      ┌─────────────────┐
      │ User Service │      │Wallet Service│      │Transaction      │
      │      Go      │      │      Go      │      │Service          │
      └──────┬───────┘      └──────┬───────┘      │Go               │
             │                     │              └────────┬────────┘
             │                     │                       │
             ▼                     ▼                       ▼
        PostgreSQL            PostgreSQL                  Kafka
                                                            │
                                      ┌─────────────────────┼───────────────────┐
                                      │                     │                   │
                                      ▼                     ▼                   ▼
                               Notification            Audit Service      Analytics
                                  Service                   Go             (future)
                                     Go
```

## 3. Repository Structure

```text
kimo-wallet/
│
├── apps/
│   ├── web/
│   │   └──                  # Next.js PWA
│   │
│   ├── api-gateway/
│   │   └──                  # Public HTTP API / BFF boundary
│   │
│   ├── user-service/
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── migrations/
│   │   ├── gen/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   ├── wallet-service/
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── migrations/
│   │   ├── gen/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   ├── transaction-service/
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── migrations/
│   │   ├── gen/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── notification-service/
│       ├── cmd/
│       ├── internal/
│       ├── gen/
│       ├── Dockerfile
│       ├── go.mod
│       └── go.sum
│
├── packages/
│   ├── contracts/
│   │   ├── user/
│   │   │   └── v1/
│   │   │       └── user.proto
│   │   ├── wallet/
│   │   │   └── v1/
│   │   │       └── wallet.proto
│   │   └── transaction/
│   │       └── v1/
│   │           └── transaction.proto
│   │
│   └── ui/
│
├── infrastructure/
│   ├── docker/
│   ├── kafka/
│   ├── postgres/
│   ├── prometheus/
│   └── grafana/
│
├── docs/
│   ├── architecture/
│   └── decisions/
│
├── buf.yaml
├── buf.gen.yaml
├── docker-compose.yml
├── Makefile
└── README.md
```

## 4. Service Responsibilities

### API Gateway

The API Gateway is the public boundary for the frontend.

Responsibilities:

- Expose HTTP/JSON APIs.
- Authenticate incoming requests.
- Validate request shape.
- Apply rate limiting.
- Translate HTTP requests into gRPC calls.
- Aggregate responses when appropriate.
- Hide internal service topology from the frontend.

The gateway should not own business logic or financial state.

```text
Next.js
   │
   │ HTTPS / JSON
   ▼
API Gateway
   │
   │ gRPC
   ▼
Internal Services
```

### User Service

Owns:

- User identity.
- Profile.
- Authentication-related data.
- Sessions/devices.
- PIN/security metadata.

It should be the source of truth for user identity.

### Wallet Service

Owns:

- Wallet.
- Currency.
- Wallet status.
- Available balance.
- Balance-related operations.

The wallet service should protect balance consistency and coordinate with transaction processing rather than allowing arbitrary balance mutation.

### Transaction Service

Owns:

- Transfer/payment requests.
- Transaction lifecycle.
- Idempotency.
- Transaction state.
- Financial operation orchestration.
- Ledger-related transaction records.

This is the core domain service.

### Notification Service

Owns:

- Notification preferences.
- In-app notifications.
- Push notification delivery.
- Notification event consumption.

It should not block the financial transaction path.

## 5. Communication Patterns

### Synchronous

Use gRPC for operations where an immediate response is required.

Examples:

```text
API Gateway → User Service
API Gateway → Wallet Service
API Gateway → Transaction Service

Transaction Service → Wallet Service
```

### Asynchronous

Use Kafka for events where processing can happen independently.

Examples:

```text
transaction.created
transaction.completed
transaction.failed
wallet.balance.updated
user.created
notification.requested
```

## 6. Protobuf Contract Strategy

Protocol Buffers are the source of truth for internal service contracts.

Contracts live under:

```text
packages/contracts/
```

Example:

```text
packages/contracts/
├── user/
│   └── v1/
│       └── user.proto
├── wallet/
│   └── v1/
│       └── wallet.proto
└── transaction/
    └── v1/
        └── transaction.proto
```

Version contracts explicitly:

```text
user/v1
user/v2
```

Avoid modifying an existing contract in a breaking way.

Use Buf for:

- Formatting.
- Linting.
- Breaking-change detection.
- Code generation.

Generated Go code belongs with the consuming service:

```text
apps/wallet-service/gen/
apps/transaction-service/gen/
```

The principle is:

> Share contracts, not service implementations.

## 7. Transaction Architecture

A transfer begins at the API Gateway:

```text
POST /transactions/transfer
        │
        ▼
API Gateway
        │
        ▼
Transaction Service
        │
        ├── Validate request
        ├── Validate idempotency key
        ├── Create transaction
        └── Publish transaction.created
                    │
                    ▼
                  Kafka
                    │
                    ▼
              Wallet Service
                    │
              ┌─────┴─────┐
              │           │
           Success       Failure
              │           │
              ▼           ▼
        Update balance   Reject/
              │          compensate
              ▼
       transaction.completed
```

The exact orchestration can evolve as the project grows. The important requirement is that the financial effect remains atomic and auditable.

## 8. Idempotency

Every financial mutation should support an idempotency key.

```text
Client
  │
  │ Idempotency-Key: abc123
  ▼
Transaction Service
  │
  ├── Does abc123 already exist?
  │
  ├── YES → return existing result
  │
  └── NO → process transaction
```

The idempotency record should be stored durably and associated with the operation result.

This prevents duplicate transfers caused by:

- User double-clicking.
- Network retries.
- Client retries.
- Gateway retries.

## 9. Concurrency Control

The system must protect against concurrent spending.

Example:

```text
Balance = Rp100.000

Request A: pay Rp80.000
Request B: pay Rp80.000
```

Only one operation can consume the available balance.

Potential implementation:

```text
BEGIN TRANSACTION

SELECT wallet
FOR UPDATE

validate balance

update wallet / ledger

COMMIT
```

The exact mechanism should be selected based on the final persistence model, but the invariant is:

> A wallet cannot successfully spend more than its available balance.

## 10. Ledger Model

The ledger should be append-oriented and auditable.

Conceptually:

```text
Ledger Entry
├── id
├── transaction_id
├── wallet_id
├── type
├── amount
├── currency
├── direction
├── created_at
└── metadata
```

For transfers, use corresponding debit/credit entries.

Example:

```text
Wallet A
  DEBIT  Rp100.000

Wallet B
  CREDIT Rp100.000
```

This provides a reliable history of financial effects.

## 11. Event-Driven Architecture

Events represent facts that have already happened.

Example:

```text
transaction.completed
```

Consumers can independently react:

```text
transaction.completed
        │
        ├── Notification Service
        │       └── Send "Transfer successful"
        │
        ├── Audit Service
        │       └── Record audit event
        │
        └── Analytics
                └── Update metrics
```

Consumers should be idempotent because Kafka delivery may result in repeated processing.

## 12. Failure Handling

The architecture should distinguish:

- Validation failure.
- Business-rule failure.
- Temporary infrastructure failure.
- Permanent processing failure.

For asynchronous processing:

```text
Event
  ↓
Consumer
  ↓
Failure
  ↓
Retry
  ↓
Retry
  ↓
Dead Letter Queue
```

Retries should use controlled backoff rather than an infinite tight loop.

## 13. Database Ownership

Each service should own its persistence model.

Preferred initial setup:

```text
User Service
   └── user database/schema

Wallet Service
   └── wallet database/schema

Transaction Service
   └── transaction database/schema
```

Services should not directly query another service's tables.

Communication across service boundaries should happen through:

- gRPC.
- Events.

For a portfolio project, separate PostgreSQL databases or schemas can be used depending on operational complexity.

The architectural rule remains:

> A service owns its data.

## 14. Caching

Redis can be introduced for:

- Session/token-related data where appropriate.
- Rate limiting.
- Frequently accessed non-authoritative data.
- Short-lived idempotency/processing state when combined with durable persistence.

Redis must not become the authoritative source of financial balance.

## 15. Frontend Architecture

The frontend is a Next.js PWA.

Suggested structure:

```text
apps/web/
├── app/
│   ├── (auth)/
│   ├── (home)/
|   |   ├── wallet/ 
│   │   ├── activity/
│   │   ├── scan/
│   │   └── profile/
│   └── ...
│
├── components/
├── features/
│   ├── auth/
│   ├── wallet/
│   ├── transaction/
│   ├── payment/
│   └── notification/
│
├── lib/
├── hooks/
├── services/
└── public/
```

Use Server Components by default and Client Components where interactivity requires them.

Potential frontend stack:

- Next.js
- TypeScript
- Tailwind CSS
- shadcn/ui
- TanStack Query
- Formik
- Yup

## 16. PWA Architecture

The web app should support:

```text
Browser
   │
   ├── Service Worker
   │      ├── Static assets
   │      └── Read-only cache
   │
   └── Next.js Application
```

Offline behavior is intentionally read-only.

Never queue a financial mutation for later execution without an explicit, carefully designed financial transaction model.

## 17. Observability

Each service should emit structured logs.

Example:

```json
{
  "level": "info",
  "service": "transaction-service",
  "trace_id": "abc123",
  "transaction_id": "txn_123",
  "event": "transaction.completed"
}
```

Metrics should include:

- Request count.
- Request latency.
- Error rate.
- Transaction success/failure rate.
- Kafka consumer lag.
- Event processing latency.
- Database connection health.

Suggested stack:

```text
Go Services
    │
    ├── OpenTelemetry
    │
    ├── Prometheus
    │
    └── Structured Logs
             │
             ▼
          Grafana
```

## 18. Distributed Tracing

A request should be traceable across services:

```text
Next.js
  ↓
API Gateway
  ↓
Transaction Service
  ↓
Wallet Service
  ↓
PostgreSQL
```

The same trace context should propagate through gRPC and asynchronous event processing where supported.

## 19. Security Architecture

Security controls should include:

- TLS in deployed environments.
- Secure authentication.
- Password/PIN hashing.
- Short-lived access tokens.
- Refresh token protection.
- Authorization checks.
- Rate limiting.
- Request validation.
- Audit logs.
- Secret management.
- Security headers.
- No sensitive information in logs.

Financial operations should require explicit authorization and, where applicable, PIN confirmation.

## 20. Deployment Architecture

### Local

Docker Compose:

```text
docker compose
├── web
├── api-gateway
├── user-service
├── wallet-service
├── transaction-service
├── notification-service
├── postgres
├── redis
├── kafka
├── prometheus
└── grafana
```

### Future cloud deployment

A possible future deployment:

```text
Users
  │
  ▼
CDN / Edge
  │
  ├──────────────► Next.js
  │
  ▼
API Gateway
  │
  ▼
Container Platform
  ├── user-service
  ├── wallet-service
  ├── transaction-service
  └── notification-service
          │
          ├── PostgreSQL
          ├── Redis
          └── Kafka
```

The cloud implementation is intentionally deferred until the local architecture is stable.

## 21. CI/CD

GitHub Actions should eventually perform:

```text
Pull Request
    │
    ├── Lint
    ├── Unit Tests
    ├── Integration Tests
    ├── Buf Lint
    ├── Buf Breaking Check
    ├── Go Build
    └── Next.js Build
           │
           ▼
        Merge
           │
           ▼
      Docker Build
           │
           ▼
        Deploy
```

## 22. Architectural Principles

1. **Services own their data.**
2. **Contracts are explicit and versioned.**
3. **Financial mutations are idempotent.**
4. **Financial state changes are auditable.**
5. **Events represent facts, not commands disguised as facts.**
6. **Consumers are idempotent.**
7. **The API Gateway does not contain domain logic.**
8. **The frontend does not directly communicate with internal services.**
9. **Offline mode is read-only for financial data.**
10. **Observability is part of the system, not an afterthought.**
11. **Start with a small number of services and split only when justified.**
12. **Prefer boring, understandable infrastructure over unnecessary complexity.**

## 23. Architecture Decision Records

Important architectural decisions should be documented under:

```text
docs/decisions/
```

Suggested ADRs:

```text
001-monorepo.md
002-grpc-and-protobuf.md
003-service-data-ownership.md
004-event-driven-transactions.md
005-idempotency.md
006-ledger-model.md
007-database-concurrency.md
008-pwa-offline-strategy.md
009-observability.md
```

These documents should explain the problem, considered alternatives, decision, and consequences.
