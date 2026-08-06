package entities

// SearchHint is a minimal entity identity for client search indexes.
type SearchHint struct {
	Title     string
	Type      string
	UniqueKey string
}

// ListSearchHints returns title/type/key for every entity instance.
// It holds only a read lock and does not copy entity Data payloads.
func ListSearchHints() []SearchHint {
	rwmutex.RLock()
	defer rwmutex.RUnlock()

	count := 0
	for _, instances := range entities {
		count += len(instances)
	}

	hints := make([]SearchHint, 0, count)
	for entityType, instances := range entities {
		for _, entity := range instances {
			if entity == nil {
				continue
			}

			hints = append(hints, SearchHint{
				Title:     entity.Title,
				Type:      entityType,
				UniqueKey: entity.UniqueKey,
			})
		}
	}

	return hints
}
