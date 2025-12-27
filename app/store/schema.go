package store

import (
	"sync"
	"time"
)

var db = make(map[string]*Value)

var mut sync.Mutex = sync.Mutex{}

type Value struct {
	StreamData *Stream
	Data       string
	ExpiresAt  *time.Time
}

// StreamEntry represents a single entry in a stream
type StreamEntry struct {
	Id     string
	Fields map[string]string
}

// Stream represents a Redis stream
type Stream struct {
	Entries       []*StreamEntry
	LastID        string
	LastTimestamp int64
	LastSequence  int64
}
