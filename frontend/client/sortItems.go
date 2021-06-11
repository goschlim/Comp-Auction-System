package client

import (
	"sort"
)

// sortItemsByDateASC sorts items by date in Slice in ASC order,
// that is, by closest date first.
func sortItemsByDateASC(items []*item) []*item {
	sort.Slice(items, func(i, j int) bool {
		return items[i].ItemClosingDate < items[j].ItemClosingDate
	})
	return items
}
