# Distributed KV Store

A distributed key-value store built in Go, designed for fault tolerance and high availability.

## Features

- **Distributed Consensus**: Leader election and log replication.
- **Write-Ahead Logging (WAL)**: Ensures data durability and state recovery across restarts.
- **HTTP API**: Simple REST-like endpoints for interacting with the key-value store. (to be replaced!)

## Getting Started


### Running Nodes

Start the first node:

```bash
go run main.go -id 1 -port 8080 -path ./tmp
```

Start additional nodes and specify the peers:

```bash
go run main.go -id 2 -port 8081 -path ./tmp -peers "1=localhost:8080"
go run main.go -id 3 -port 8082 -path ./tmp -peers "1=localhost:8080,2=localhost:8081"
```

*Note that nodes will automatically run elections and agree on a consensus leader.*

## HTTP API

The API exposes endpoints to manage key-value pairs:

- **Get a value:** `GET /value?key=<key>`
- **Put a value:** `PUT /value` 
  - Body: `{"key": "your_key", "value": "your_value"}`
  - *Writes must be directed to the current leader.*
- **Delete a value:** `DELETE /value?key=<key>`
  - *Deletes must be directed to the current leader.*

## Upcoming Features

- **Consistent Hashing**: To distribute keys more smoothly and efficiently across the cluster, avoiding major redistributions when nodes are added or removed.
- **Data Partitioning**: Splitting the dataset into multiple partitions to scale horizontal read and write capacity beyond a single node's limits.
