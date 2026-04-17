# Distributed KV Store

A distributed key-value store built in Go, with Raft-based replication per shard.

## Architecture and Design


At its core, this system leverages the **Raft consensus algorithm** to guarantee distributed agreement across the cluster. To optimize performance without the overhead of a traditional database engine, the key-value store operates entirely in-memory. Data durability and fault tolerance are preserved by maintaining an immutable, sequential ledger of state transitions known as the Write-Ahead Log (WAL).

Keys are currently mapped to shards via static hashing (`fnv32a(key) % shardCount`), and each shard runs an independent Raft group. A single node can host multiple shard stores (one per local shard).

**Core Mechanisms:**
1. **Shard Selection:** For each request, the key is mapped to a shard using static hashing. If the shard is not local to the receiving node, the request is redirected to a node that hosts that shard.
2. **Leader Election (per shard):** Every shard store initializes as a `Follower`. If a follower fails to receive a heartbeat (`AppendEntries`) within a randomized timeout window, it transitions to `Candidate` and starts election for that shard.
3. **Log Replication (per shard):** Write operations (`PUT`, `DELETE`) are committed by the leader of the target shard. The leader appends to WAL and replicates via `AppendEntries` gRPC calls.
4. **Commit & Application:** A log entry is committed after quorum replication within that shard's Raft group, then applied to the in-memory map.
5. **Crash Recovery:** On restart, each local shard store rebuilds state by replaying its WAL from disk.

### System Components
- **API Engine:** Handles HTTP requests, resolves target shard from key, and redirects when request lands on a non-owner node or non-leader replica.
- **Shard Manager:** Maintains shard topology (`shardID -> nodes`) and local shard stores.
- **State Machine:** A highly optimized in-memory map holding the operational key-value records.
- **Consensus Module:** Runs Raft independently per shard, including elections and replication.
- **Write-Ahead Log (WAL):** A reliable append-only persistence layer under each shard path.

## Codebase Overview

- `main.go`: Entry point. Parses Node ID, ports, data path, and `shardMap` to initialize the node.
- `internal/api/`:
  - `handlers.go`: Exposes REST endpoints (`GET`, `PUT`, `DELETE`). Routes by shard first, then ensures writes are handled by shard leader.
  - `shard-utils.go`: Shard manager and static-hash shard resolution (`fnv32a`).
  - `raft-rpc-server.go`: gRPC server that dispatches Raft RPCs (`AppendEntries`, `RequestVote`) to the correct local shard store.
  - `router.go` & `server.go`: Composes the HTTP request multiplexer and initializes the HTTP daemon.
- `internal/proto/raft/`: Hosts the gRPC Protocol Buffers definitions alongside auto-generated Go bindings. Defines Raft's fundamental RPCs: `AppendEntries` and `RequestVote`.
- `internal/server/store/`: The domain core that bridges the consensus algorithm with the data storage layer.
  - `store.go`: The primary struct encompassing the node's Raft state, WAL subsystem, and the in-memory map. Governs the promotion of committed entries into the state machine.
  - `raft.go`: Handles processing logic for incoming `AppendEntries` and `RequestVote` RPC invocations. Explicitly enforces cluster consistency and orchestrates leader election timer mechanics.
  - `followers.go`: Facilitates the outward synchronization process, allowing the leader to proactively stream and replicate log entries out to the follower node pool.
- `internal/server/wal/`:
  - `wal.go`: A custom Write-Ahead Log engine. Manages precise binary serialization, atomic disk appends, log recovery reads, and targeted truncations to reliably ensure data persistence.

## Features

- **Sharding with static hashing**: Keys are distributed to shards using `fnv32a(key) % shardCount`.
- **Multi-shard per node**: A node can run multiple local Raft-backed shard stores.
- **Distributed consensus (per shard)**: Leader election and log replication for each shard group.
- **Write-Ahead Logging (WAL)**: Ensures durability and recovery on restarts.
- **HTTP API**: REST endpoints for reads/writes/deletes with automatic redirects.

## Getting Started


### Running Nodes

Start the first node:

```bash
go run main.go -id 1 -port 8080 -path ./tmp -shardMap "0=1@localhost:8080,2@localhost:8081,3@localhost:8082;1=1@localhost:8080,2@localhost:8081,3@localhost:8082"
```

Start additional nodes with the same `-shardMap`:

```bash
go run main.go -id 2 -port 8081 -path ./tmp -shardMap "0=1@localhost:8080,2@localhost:8081,3@localhost:8082;1=1@localhost:8080,2@localhost:8081,3@localhost:8082"
go run main.go -id 3 -port 8082 -path ./tmp -shardMap "0=1@localhost:8080,2@localhost:8081,3@localhost:8082;1=1@localhost:8080,2@localhost:8081,3@localhost:8082"
```

`-shardMap` format:

`shardID=nodeID@host:port,nodeID@host:port;shardID=nodeID@host:port,...`

Each node also opens a Raft gRPC listener on `http_port + 10000`.

## HTTP API

The API exposes endpoints to manage key-value pairs:

- **Get a value:** `GET /api/v1/value?key=<key>`
- **Put a value:** `PUT /api/v1/value` 
  - Body: `{"key": "your_key", "value": "your_value"}`
  - If request hits the wrong node/shard or a follower, it is redirected.
- **Delete a value:** `DELETE /api/v1/value?key=<key>`
  - If request hits the wrong node/shard or a follower, it is redirected.

## Upcoming Features

- **Consistent Hashing**: Replace static modulo hashing to reduce key movement during topology changes.
