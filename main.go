package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultPort = "6379"
)

// DataType represents the type of data stored
type DataType int

const (
	TypeString DataType = iota
	TypeStream
)

// StreamEntry represents a single entry in a stream
type StreamEntry struct {
	id     string
	fields map[string]string
}

// Stream represents a Redis stream
type Stream struct {
	entries       []*StreamEntry
	lastID        string
	lastTimestamp int64
	lastSequence  int64
}

// Value holds the data and optional expiry time
type Value struct {
	dataType   DataType
	stringData string
	streamData *Stream
	expiry     *time.Time // nil means no expiry
}

// Store holds the key-value pairs
type Store struct {
	mu   sync.RWMutex
	data map[string]*Value
}

var store = &Store{
	data: make(map[string]*Value),
}

func main() {
	listener, err := net.Listen("tcp", ":"+DefaultPort)
	if err != nil {
		fmt.Printf("Failed to bind to port %s: %v\n", DefaultPort, err)
		return
	}
	defer listener.Close()

	fmt.Printf("Redis server listening on port %s\n", DefaultPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Read RESP command
		command, err := readRESP(reader)
		if err != nil {
			return
		}

		// Process command
		response := processCommand(command)

		// Send response
		conn.Write([]byte(response))
	}
}

func readRESP(reader *bufio.Reader) ([]string, error) {
	// Read array size line (*<count>\r\n)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	var args []string

	// Simple parsing for RESP arrays
	if len(line) > 0 && line[0] == '*' {
		var count int
		fmt.Sscanf(line, "*%d", &count)

		// Parse each array element
		for i := 0; i < count; i++ {
			// Read bulk string length line ($<length>\r\n)
			_, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}

			// Read the actual string
			str, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}

			args = append(args, strings.TrimSpace(str))
		}
	}

	return args, nil
}

func processCommand(args []string) string {
	if len(args) == 0 {
		return "-ERR no command provided\r\n"
	}

	command := strings.ToUpper(args[0])

	switch command {
	case "PING":
		return handlePing(args)
	case "ECHO":
		return handleEcho(args)
	case "SET":
		return handleSet(args)
	case "GET":
		return handleGet(args)
	case "TYPE":
		return handleType(args)
	case "XADD":
		return handleXAdd(args)
	default:
		return fmt.Sprintf("-ERR unknown command '%s'\r\n", args[0])
	}
}

func handlePing(args []string) string {
	if len(args) == 1 {
		// PING without arguments returns "PONG"
		return "+PONG\r\n"
	}
	// PING with message returns the message
	return fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])
}

func handleEcho(args []string) string {
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'echo' command\r\n"
	}
	// ECHO returns the message as a bulk string
	return fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])
}

func handleSet(args []string) string {
	if len(args) < 3 {
		return "-ERR wrong number of arguments for 'set' command\r\n"
	}

	key := args[1]
	value := args[2]
	var expiry time.Duration

	// Parse optional EX/PX flags
	for i := 3; i < len(args); i++ {
		option := strings.ToUpper(args[i])
		switch option {
		case "EX":
			if i+1 >= len(args) {
				return "-ERR syntax error\r\n"
			}
			seconds, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil || seconds <= 0 {
				return "-ERR invalid expire time in 'set' command\r\n"
			}
			expiry = time.Duration(seconds) * time.Second
			i++ // Skip next arg
		case "PX":
			if i+1 >= len(args) {
				return "-ERR syntax error\r\n"
			}
			milliseconds, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil || milliseconds <= 0 {
				return "-ERR invalid expire time in 'set' command\r\n"
			}
			expiry = time.Duration(milliseconds) * time.Millisecond
			i++ // Skip next arg
		}
	}

	store.mu.Lock()
	val := &Value{
		dataType:   TypeString,
		stringData: value,
	}
	if expiry > 0 {
		expireTime := time.Now().Add(expiry)
		val.expiry = &expireTime
	}
	store.data[key] = val
	store.mu.Unlock()

	return "+OK\r\n"
}

func handleGet(args []string) string {
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'get' command\r\n"
	}

	key := args[1]

	store.mu.Lock()
	defer store.mu.Unlock()

	val, exists := store.data[key]
	if !exists {
		return "$-1\r\n" // Null bulk string
	}

	// Check if key has expired
	if val.expiry != nil && time.Now().After(*val.expiry) {
		// Key has expired, delete it
		delete(store.data, key)
		return "$-1\r\n" // Null bulk string
	}

	// Check if it's a string type
	if val.dataType != TypeString {
		return "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(val.stringData), val.stringData)
}

func handleType(args []string) string {
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'type' command\r\n"
	}

	key := args[1]

	store.mu.RLock()
	val, exists := store.data[key]
	store.mu.RUnlock()

	if !exists {
		return "+none\r\n"
	}

	// Check if key has expired
	if val.expiry != nil && time.Now().After(*val.expiry) {
		return "+none\r\n"
	}

	// Return the data type
	switch val.dataType {
	case TypeString:
		return "+string\r\n"
	case TypeStream:
		return "+stream\r\n"
	default:
		return "+none\r\n"
	}
}

func handleXAdd(args []string) string {
	// XADD key ID field value [field value ...]
	if len(args) < 4 {
		return "-ERR wrong number of arguments for 'xadd' command\r\n"
	}

	// Must have an even number of field-value pairs
	if (len(args)-3)%2 != 0 {
		return "-ERR wrong number of arguments for 'xadd' command\r\n"
	}

	key := args[1]
	idArg := args[2]

	// Parse field-value pairs
	fields := make(map[string]string)
	for i := 3; i < len(args); i += 2 {
		fields[args[i]] = args[i+1]
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	// Get or create stream
	var stream *Stream
	val, exists := store.data[key]

	if exists {
		// Check if it's a stream type
		if val.dataType != TypeStream {
			return "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
		}
		stream = val.streamData
	} else {
		// Create new stream
		stream = &Stream{
			entries:       make([]*StreamEntry, 0),
			lastTimestamp: 0,
			lastSequence:  0,
		}
		val = &Value{
			dataType:   TypeStream,
			streamData: stream,
		}
		store.data[key] = val
	}

	// Generate or validate ID
	var entryID string
	if idArg == "*" {
		// Auto-generate ID using current timestamp
		timestamp := time.Now().UnixMilli()
		sequence := int64(0)

		if timestamp == stream.lastTimestamp {
			sequence = stream.lastSequence + 1
		} else if timestamp < stream.lastTimestamp {
			// Use last timestamp + 1
			timestamp = stream.lastTimestamp
			sequence = stream.lastSequence + 1
		}

		stream.lastTimestamp = timestamp
		stream.lastSequence = sequence
		entryID = fmt.Sprintf("%d-%d", timestamp, sequence)
	} else {
		// Use provided ID
		// Basic validation - should be in format timestamp-sequence
		parts := strings.Split(idArg, "-")
		if len(parts) != 2 {
			return "-ERR Invalid stream ID specified as stream command argument\r\n"
		}

		timestamp, err1 := strconv.ParseInt(parts[0], 10, 64)
		sequence, err2 := strconv.ParseInt(parts[1], 10, 64)

		if err1 != nil || err2 != nil {
			return "-ERR Invalid stream ID specified as stream command argument\r\n"
		}

		// Check if ID is greater than last ID
		if len(stream.entries) > 0 {
			if timestamp < stream.lastTimestamp || (timestamp == stream.lastTimestamp && sequence <= stream.lastSequence) {
				return "-ERR The ID specified in XADD is equal or smaller than the target stream top item\r\n"
			}
		}

		stream.lastTimestamp = timestamp
		stream.lastSequence = sequence
		entryID = idArg
	}

	// Create and add entry
	entry := &StreamEntry{
		id:     entryID,
		fields: fields,
	}
	stream.entries = append(stream.entries, entry)

	// Return the generated/used ID as bulk string
	return fmt.Sprintf("$%d\r\n%s\r\n", len(entryID), entryID)
}
