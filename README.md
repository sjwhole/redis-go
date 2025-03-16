# Go Redis Server Implementation

A lightweight Redis server implementation in Go that supports basic Redis commands and persistence via RDB files.

## Table of Contents

* [Features](#features)
* [Components Overview](#components-overview)
* [Getting Started](#getting-started)
* [Usage Examples](#usage-examples)
* [References](#references)

## Features

* Implementation of core Redis commands (PING, ECHO, SET, GET, CONFIG GET, KEYS, INFO)
* Key-value storage with expiration support
* RDB file format persistence
* Command parsing based on the Redis protocol
* Support for master-slave replication configuration

## Components Overview

1. Parser
    * **Algorithm**: Redis Protocol (RESP) parser
    * Converts client messages into executable commands
    * Handles command arguments and parsing
    * Supports different command formats and arguments

2. Command System
    * Command interface for uniform execution
    * Individual command implementations (PingCommand, EchoCommand, etc.)
    * Error handling and response formatting according to RESP

3. Storage
    * In-memory key-value database implementation
    * Support for key expiration (both milliseconds and seconds)
    * Concurrent access handling with timers for expiry

4. Persistence
    * **Algorithm**: RDB file format parsing
    * Reading and parsing RDB dump files
    * Support for different value types and expiry formats
    * Binary data handling for compact storage

## Getting Started

### Prerequisites

* Go 1.16+

### Installation

```
git clone https://github.com/sjwhole/redis-go.git
cd redis-go
```

## Usage Examples

### Starting the Redis Server

```
chmod +x go-redis-server.sh
./spawn-redis-server.sh
```

### With RDB File Loading

```
./spawn-redis-server.sh --dbfilename dump.rdb
```

### Connecting with Redis CLI

```
redis-cli -p 6379
> PING
PONG
> SET mykey "Hello World"
OK
> GET mykey
"Hello World"
> SET key-with-expiry "I will expire" PX 5000
OK
> KEYS *
1) "mykey"
2) "key-with-expiry"
```

## References

* [Code Crafters - Build Your Own Redis](https://codecrafters.io/redis)
* [Redis Protocol Specification](https://redis.io/topics/protocol)
* [Redis Command Reference](https://redis.io/commands)
* [Redis RDB File Format](https://github.com/sripathikrishnan/redis-rdb-tools/wiki/Redis-RDB-Dump-File-Format)
