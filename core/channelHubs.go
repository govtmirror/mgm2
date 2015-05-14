package core

// HostHub contains the host related channels to allow for easy passing
type HostHub struct {
	HostStatsNotifier chan HostStats
	HostNotifier      chan Host
}

type sessionLookup struct {
	jobLink      chan Job
	hostStatLink chan HostStats
	hostLink     chan Host
	accessLevel  uint8
}
