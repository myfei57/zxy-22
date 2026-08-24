// Package reading: threshold evaluation of a single reading.
package reading

// Evaluate reports whether the reading value exceeds the given threshold.
func Evaluate(rd Reading, threshold float64) bool {
	return rd.Value > threshold
}
