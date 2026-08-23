// Package audit: event statistics.
package audit

// Stats counts events by type.
func Stats(events []Event) map[string]int {
	counts := map[string]int{}
	for _, event := range events {
		counts[event.Type]++
	}
	return counts
}
