package entities

type SearchHint struct {
	Title     string
	Type      string
	UniqueKey string
}

func ListSearchHints() []SearchHint {
	rwmutex.RLock()
	defer rwmutex.RUnlock()

	hints := make([]SearchHint, 0)
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
