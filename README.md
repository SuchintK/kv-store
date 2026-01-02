# GoDis Key-Value Store

A key-value store implementation in Go, supporting multiple data structures and advanced features.

## Quick Start

Start a local server:
```bash
./spawn_redis_server.sh
```

This starts a Redis server on port **6379** by default.

Start a replica server:
```bash
./spawn_redis_server.sh --port 3000 --replicaof localhost 6379
```

Test the server:
```bash
$ redis-cli ping
PONG
```

---

## Features

### Basic Commands

#### PING
Check server connectivity.
```bash
PING
# Returns: PONG
```

#### ECHO
Echo the given string.
```bash
ECHO "Hello World"
# Returns: "Hello World"
```

---

### String Operations

#### SET
Set key to hold the string value with optional expiration.
```bash
SET mykey "value"
SET session "data" PX 5000  # Expire in 5000ms
```

#### GET
Get the value of a key.
```bash
GET mykey
# Returns: "value"
```

#### INCR
Increment the integer value of a key by one.
```bash
INCR counter
# Returns: 1
```

---

### List Operations

Lists are ordered collections of strings. Elements can be added/removed from both ends.

#### LPUSH
Insert elements at the head of a list.
```bash
LPUSH mylist "element1" "element2"
# Returns: 2
```

#### RPUSH
Append elements to the tail of a list.
```bash
RPUSH mylist "element3" "element4"
# Returns: 4
```

#### LPOP
Remove and return the first element of a list.
```bash
LPOP mylist
# Returns: "element2"
```

#### RPOP
Remove and return the last element of a list.
```bash
RPOP mylist
# Returns: "element4"
```

#### LLEN
Get the length of a list.
```bash
LLEN mylist
# Returns: 2
```

#### LRANGE
Get a range of elements from a list.
```bash
LRANGE mylist 0 -1  # Get all elements
# Returns: ["element1", "element3"]
```

#### BLPOP
Blocking version of LPOP. Waits for an element to become available.
```bash
BLPOP mylist 5  # Wait up to 5 seconds
# Returns: ["mylist", "element1"]
```

---

### Sorted Set Operations

Sorted sets store unique members with associated scores, automatically sorted by score.

#### ZADD
Add members with scores to a sorted set.
```bash
ZADD leaderboard 100 "player1" 200 "player2"
# Returns: 2
```

#### ZCARD
Get the number of members in a sorted set.
```bash
ZCARD leaderboard
# Returns: 2
```

#### ZRANK
Get the rank (index) of a member in a sorted set.
```bash
ZRANK leaderboard "player1"
# Returns: 0
```

#### ZSCORE
Get the score of a member in a sorted set.
```bash
ZSCORE leaderboard "player1"
# Returns: "100"
```

#### ZRANGE
Get members in a score range.
```bash
ZRANGE leaderboard 0 -1 WITHSCORES
# Returns: ["player1", "100", "player2", "200"]
```

#### ZREM
Remove members from a sorted set.
```bash
ZREM leaderboard "player1"
# Returns: 1
```

---

### Geospatial Operations

Geospatial indexes store locations with latitude and longitude coordinates, enabling radius queries and distance calculations.

#### GEOADD
Add geospatial locations (longitude, latitude, member) to a key.
```bash
GEOADD locations 13.361389 38.115556 "Palermo"
# Returns: 1

GEOADD locations 15.087269 37.502669 "Catania" 12.5 37.8 "Agrigento"
# Returns: 2
```

**Validation:**
- Longitude: -180.0 to 180.0
- Latitude: -85.05112878 to 85.05112878

#### GEOPOS
Get the coordinates (longitude, latitude) of members.
```bash
GEOPOS locations "Palermo" "Catania"
# Returns: [[13.361389, 38.115556], [15.087269, 37.502669]]

GEOPOS locations "NonExistent"
# Returns: [nil]
```

#### GEODIST
Calculate the distance between two members.
```bash
GEODIST locations "Palermo" "Catania"
# Returns: "166274.1516" (meters by default)

GEODIST locations "Palermo" "Catania" km
# Returns: "166.2741516"

GEODIST locations "Palermo" "Catania" mi
# Returns: "103.3182"
```

**Supported units:**
- `m` - meters (default)
- `km` - kilometers
- `mi` - miles
- `ft` - feet

#### GEORADIUS
Query members within a radius from coordinates.
```bash
GEORADIUS locations 15.0 37.0 200 km
# Returns: ["Palermo", "Catania"]

GEORADIUS locations 15.0 37.0 100 km
# Returns: ["Catania"]
```

**Optional flags:**
```bash
# Include distances
GEORADIUS locations 15.0 37.0 200 km WITHDIST
# Returns: [["Catania", "56.4413"], ["Palermo", "190.4424"]]

# Include coordinates
GEORADIUS locations 15.0 37.0 200 km WITHCOORD
# Returns: [["Palermo", [13.361389, 38.115556]], ...]

# Include geohash
GEORADIUS locations 15.0 37.0 200 km WITHHASH
# Returns: [["Palermo", 3479099956230698], ...]

# Limit results
GEORADIUS locations 15.0 37.0 200 km COUNT 1
# Returns: ["Catania"]

# Sort by distance
GEORADIUS locations 15.0 37.0 200 km ASC
# Returns: ["Catania", "Palermo"] (closest first)

GEORADIUS locations 15.0 37.0 200 km DESC
# Returns: ["Palermo", "Catania"] (farthest first)

# Combine options
GEORADIUS locations 15.0 37.0 200 km WITHDIST WITHCOORD ASC COUNT 2
```

**Implementation details:**
- Uses 52-bit geohash encoding (26-bit latitude + 26-bit longitude)
- Haversine formula for distance calculations
- Earth radius: 6372797.560856 meters
- Stored in sorted sets with geohash as score

---

### Stream Operations

Streams are append-only log data structures for event streaming.

#### XADD
Append a new entry to a stream.
```bash
XADD mystream * field1 value1 field2 value2
# Returns: "1640000000000-0"
```

#### XRANGE
Query a range of entries in a stream.
```bash
XRANGE mystream - +
# Returns all entries
```

#### XREAD
Read entries from one or more streams.
```bash
XREAD STREAMS mystream 0
# Returns new entries since ID 0
```

---

### Pub/Sub

Publish/Subscribe messaging pattern for real-time communication.

#### PUBLISH
Post a message to a channel.
```bash
PUBLISH news "Breaking news!"
# Returns: 1 (number of subscribers)
```

#### SUBSCRIBE
Subscribe to channels.
```bash
SUBSCRIBE news sports
# Receives messages from news and sports channels
```

#### UNSUBSCRIBE
Unsubscribe from channels.
```bash
UNSUBSCRIBE news
# Stops receiving messages from news channel
```

---

### Transactions

Execute multiple commands atomically.

#### MULTI
Mark the start of a transaction block.
```bash
MULTI
```

#### EXEC
Execute all queued commands in the transaction.
```bash
EXEC
# Executes all commands queued since MULTI
```

#### DISCARD
Discard all commands in the transaction queue.
```bash
DISCARD
# Cancels the transaction
```

---

### Replication

Master-replica replication for high availability and read scaling.

#### INFO
Get server information including replication status.
```bash
INFO replication
# Returns: role:master, connected_slaves:1, ...
```

#### REPLCONF
Internal command used by replicas to configure replication.
```bash
REPLCONF listening-port 6380
```

#### PSYNC
Internal command used by replicas to synchronize with master.
```bash
PSYNC ? -1
# Initiates full synchronization
```

---

## Architecture

- **RESP Protocol**: Redis Serialization Protocol for client-server communication
- **Skip List**: Efficient sorted set implementation with O(log n) operations
- **Geospatial Index**: 52-bit geohash encoding with Haversine distance calculations
- **Pub/Sub**: In-memory message broker with channel subscriptions
- **Replication**: Asynchronous master-replica data synchronization
- **Transactions**: ACID-compliant transaction support with command queuing

---

## Testing

Run all tests:
```bash
go test ./tests/...
```

Run specific test:
```bash
go test ./tests/zadd_test.go -v
```

---

## Implementation Details

- Written in Go for performance and concurrency
- Custom RESP protocol parser
- Skip list data structure for sorted sets
- Non-blocking pub/sub with goroutines
- Polling-based blocking operations (BLPOP)
- Thread-safe operations with mutex locks
