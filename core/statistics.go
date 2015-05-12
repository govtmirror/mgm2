package core

// HostStats holds mgm host statistical info
type HostStats struct {
	CPUPercent []float64
	MEMTotal   uint64
	MEMUsed    uint64
	MEMPercent float64
}
