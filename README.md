## Outboxy – Outbox Pattern Demo (Orders + Dispatcher)

This project is a minimal Go implementation of the **outbox pattern** for an e‑commerce style engine:

- `orders`: writes orders and corresponding outbox records in a single DB transaction.
- `dispatcher`: polls the outbox table and publishes events to **Redis pub/sub**.

Both services are plain Go binaries in this module.

---

## Components

- **Postgres** – primary data store for orders and outbox messages.
- **Redis** – pub/sub transport; dispatcher publishes `orders.created` events.
- **Orders service (`./orders`)**
  - Inserts into `orders` table.
  - Inserts a serialized `OrderEvent` into `outbox` table with `state = 'pending'`.
- **Dispatcher service (`./dispatcher`)**
  - Periodically (every second) selects one `pending` outbox row with `FOR UPDATE SKIP LOCKED`.
  - Publishes the message to Redis channel = `topic`.
  - Marks the outbox row as `processed`.

---

## Database schema

Example Postgres schema that matches the code:

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE orders (
  id UUID PRIMARY KEY,
  product TEXT NOT NULL,
  quantity INT NOT NULL
);

CREATE TABLE outbox (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  topic TEXT NOT NULL,
  message BYTEA NOT NULL,
  state TEXT NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  processed_at TIMESTAMPTZ
);
```

> `orders/main.go` writes to both `orders` and `outbox` inside a single transaction.

---

## Environment variables

- **`DATABASE_URL`** (required for both services)  
  Example:

  ```bash
  export DATABASE_URL="postgres://user:password@localhost:5432/outboxy?sslmode=disable"
  ```

- **`REDIS_URL`** (optional for `dispatcher`)  
  If empty, the dispatcher uses `localhost:6379`:

  ```bash
  export REDIS_URL="redis://localhost:6379"
  ```

---

## Makefile targets

From the project root (`/home/atharvamhaske/Projects/outboxy`):

- **Build both binaries**

  ```bash
  make build
  # or individually
  make build-orders
  make build-dispatcher
  ```

- **Run tests (Ginkgo/Gomega based)**

  ```bash
  make test
  ```

- **Run services (with env vars set)**

  ```bash
  # Orders service – creates an order and an outbox row
  make run-orders

  # Dispatcher – polls outbox and publishes to Redis
  make run-dispatcher
  ```

- **Go module tidy**

  ```bash
  make tidy
  ```

---

## Running everything locally

1. **Start Postgres and Redis**

   - Postgres database `outboxy` with the schema above.
   - Redis server (default: `localhost:6379`).

2. **Set environment variables**

   ```bash
   export DATABASE_URL="postgres://user:password@localhost:5432/outboxy?sslmode=disable"
   export REDIS_URL="redis://localhost:6379"
   ```

3. **Run dispatcher**

   ```bash
   make run-dispatcher
   ```

4. **Subscribe to Redis channel to see events**

   ```bash
   redis-cli
   SUBSCRIBE orders.created
   ```

5. **Trigger an order + outbox entry**

   ```bash
   make run-orders
   ```

6. **Observe**

   - Dispatcher logs: publishing + marking outbox rows as processed.
   - `redis-cli` subscriber: receives `orders.created` JSON payloads.
   - DB:
     - `orders` contains the new order.
     - `outbox` row moves from `pending` to `processed`.

---

## High‑level flow (outbox pattern)

```mermaid
flowchart LR
    A["Client / Checkout UI"] --> B["Orders service"]

    B --> C[Begin DB transaction]
    C --> D[Insert into orders table]
    C --> E[Insert into outbox table (pending)]

    D --> F["Postgres orders table"]
    E --> G["Postgres outbox table - pending"]

    H["Dispatcher service"] --> I["Every 1s select pending outbox row with FOR UPDATE SKIP LOCKED LIMIT 1"]

    I -->|row found| J[Scan row into OutboxMsg]
    J --> K["Publish message to Redis channel"]
    K --> L["Redis PubSub channel orders_created"]
    L --> M["Downstream consumers: inventory, email, analytics"]

    J --> N[Update outbox row to processed]
    N --> O["Postgres outbox table - processed"]

    I -->|no row| P["No-op, wait for next tick"]
```


