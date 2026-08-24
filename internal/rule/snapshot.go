// Package rule: snapshot inspection for the console.
package rule

// Snapshot returns the threshold snapshot the manager loaded with.
func Snapshot(manager *VersionManager, metricID string) (Threshold, error) {
	return manager.Session(metricID)
}
