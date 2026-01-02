package store

// SortedSet represents a Redis sorted set with map + skip list hybrid
type SortedSet struct {
	dict     map[string]float64 // member -> score for O(1) lookup
	skipList *SkipList          // skip list for ordered operations
}

// NewSortedSet creates a new sorted set
func NewSortedSet() *SortedSet {
	return &SortedSet{
		dict:     make(map[string]float64),
		skipList: NewSkipList(),
	}
}

// Add adds or updates a member with a score
// Returns true if member was added, false if score was updated
func (zs *SortedSet) Add(score float64, member string) bool {
	oldScore, exists := zs.dict[member]

	if exists {
		if oldScore == score {
			return false // No change
		}
		// Remove old entry from skip list
		zs.skipList.Delete(oldScore, member)
	}

	// Add/update in dict
	zs.dict[member] = score

	// Add to skip list
	zs.skipList.Insert(score, member)

	return !exists
}

// Remove removes a member from the sorted set
// Returns true if member was removed
func (zs *SortedSet) Remove(member string) bool {
	score, exists := zs.dict[member]
	if !exists {
		return false
	}

	delete(zs.dict, member)
	zs.skipList.Delete(score, member)
	return true
}

// GetScore returns the score of a member
func (zs *SortedSet) GetScore(member string) (float64, bool) {
	score, exists := zs.dict[member]
	return score, exists
}

// Card returns the cardinality (number of elements)
func (zs *SortedSet) Card() int {
	return len(zs.dict)
}

// GetRank returns the rank (0-based) of a member
func (zs *SortedSet) GetRank(member string) int {
	score, exists := zs.dict[member]
	if !exists {
		return -1
	}
	return zs.skipList.GetRank(score, member)
}

// GetRange returns members in the given rank range
func (zs *SortedSet) GetRange(start, stop int) []string {
	nodes := zs.skipList.GetRange(start, stop)
	result := make([]string, len(nodes))
	for i, node := range nodes {
		result[i] = node.Member
	}
	return result
}

// GetRangeWithScores returns members with scores in the given rank range
func (zs *SortedSet) GetRangeWithScores(start, stop int) []struct {
	Member string
	Score  float64
} {
	nodes := zs.skipList.GetRange(start, stop)
	result := make([]struct {
		Member string
		Score  float64
	}, len(nodes))
	for i, node := range nodes {
		result[i] = struct {
			Member string
			Score  float64
		}{
			Member: node.Member,
			Score:  node.Score,
		}
	}
	return result
}
