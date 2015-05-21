package core

import "github.com/M-O-S-E-S/mgm/mgm"

// HostHub contains the host related channels to allow for easy passing
type HostHub struct {
	HostStatsNotifier chan mgm.HostStat
	HostNotifier      chan mgm.Host
}

type sessionLookup struct {
	jobLink      chan mgm.Job
	hostStatLink chan mgm.HostStat
	hostLink     chan mgm.Host
	accessLevel  uint8
}
