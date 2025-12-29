package store

import (
	"math/rand"
)

const (
	maxLevel    = 32
	probability = 0.25
)

// SkipListNode represents a node in the skip list
type SkipListNode struct {
	Member string
	Score  float64
	Next   []*SkipListNode // Contains the next node
}

// SkipList represents a skip list for sorted set
type SkipList struct {
	header *SkipListNode
	level  int
	length int
}

// NewSkipList creates a new skip list
func NewSkipList() *SkipList {
	return &SkipList{
		header: &SkipListNode{
			Next: make([]*SkipListNode, maxLevel),
		},
		level:  0,
		length: 0,
	}
}

// randomLevel generates a random level for a new node
func (sl *SkipList) randomLevel() int {
	level := 0
	for level < maxLevel-1 && rand.Float64() < probability {
		level++
	}
	return level
}

// Insert inserts or updates a member with a score
func (sl *SkipList) Insert(score float64, member string) *SkipListNode {
	update := make([]*SkipListNode, maxLevel)
	current := sl.header

	// Find the position to insert
	for i := sl.level; i >= 0; i-- {
		for current.Next[i] != nil &&
			(current.Next[i].Score < score ||
				(current.Next[i].Score == score && current.Next[i].Member < member)) {
			current = current.Next[i]
		}
		update[i] = current
	}

	// Check if member already exists at this position
	current = current.Next[0]
	if current != nil && current.Score == score && current.Member == member {
		return current
	}

	// Generate random level for new node
	newLevel := sl.randomLevel()

	// Levels will be added to the top of the list
	if newLevel > sl.level {
		for i := sl.level + 1; i <= newLevel; i++ {
			update[i] = sl.header
		}
		sl.level = newLevel
	}

	// Create new node
	newNode := &SkipListNode{
		Member: member,
		Score:  score,
		Next:   make([]*SkipListNode, newLevel+1),
	}

	// Insert node
	for i := 0; i <= newLevel; i++ {
		newNode.Next[i] = update[i].Next[i]
		update[i].Next[i] = newNode
	}

	sl.length++
	return newNode
}

// Delete removes a member from the skip list
func (sl *SkipList) Delete(score float64, member string) bool {
	update := make([]*SkipListNode, maxLevel)
	current := sl.header

	// Find the node to delete
	for i := sl.level; i >= 0; i-- {
		for current.Next[i] != nil &&
			(current.Next[i].Score < score ||
				(current.Next[i].Score == score && current.Next[i].Member < member)) {
			current = current.Next[i]
		}
		update[i] = current
	}

	current = current.Next[0]
	if current == nil || current.Score != score || current.Member != member {
		return false
	}

	// Remove node from all levels
	for i := 0; i <= sl.level; i++ {
		if update[i].Next[i] != current {
			break
		}
		update[i].Next[i] = current.Next[i]
	}

	// Update level if needed
	for sl.level > 0 && sl.header.Next[sl.level] == nil {
		sl.level--
	}

	sl.length--
	return true
}

// GetRank returns the rank (0-based) of a member
// Traverses level 0 directly to count position
func (sl *SkipList) GetRank(score float64, member string) int {
	rank := 0
	current := sl.header.Next[0] // Start at first node in level 0

	// Traverse level 0 and count nodes until we find the target
	for current != nil {
		if current.Score == score && current.Member == member {
			return rank // Found it
		}
		// If we've passed where it should be, it doesn't exist
		if current.Score > score || (current.Score == score && current.Member > member) {
			return -1
		}
		rank++
		current = current.Next[0]
	}

	return -1 // Not found
}

// GetRange returns members in the given rank range [start, stop] (0-based, inclusive)
func (sl *SkipList) GetRange(start, stop int) []*SkipListNode {
	if start < 0 {
		start = sl.length + start
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = sl.length + stop
		if stop < 0 {
			stop = 0
		}
	}
	if start > stop || start >= sl.length {
		return []*SkipListNode{}
	}
	if stop >= sl.length {
		stop = sl.length - 1
	}

	result := make([]*SkipListNode, 0, stop-start+1)
	current := sl.header.Next[0]

	// Skip to start position
	for i := 0; i < start && current != nil; i++ {
		current = current.Next[0]
	}

	// Collect nodes from start to stop
	for i := start; i <= stop && current != nil; i++ {
		result = append(result, current)
		current = current.Next[0]
	}

	return result
}

// Length returns the number of elements in the skip list
func (sl *SkipList) Length() int {
	return sl.length
}
