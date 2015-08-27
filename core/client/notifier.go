package client

import "github.com/m-o-s-e-s/mgm/mgm"

//Notifier hold a bunch of channels for communicating with user sessions
type Notifier struct {
	hUp   chan mgm.Host
	hDel  chan int64
	hStat chan mgm.HostStat
	rUp   chan mgm.Region
	rDel  chan mgm.Region
	rStat chan mgm.RegionStat
	eUp   chan mgm.Estate
	eDel  chan mgm.Estate
	jUp   chan mgm.Job
	jDel  chan mgm.Job
}

//NewNotifier constructs a Notifier, initializing all internal data structures
func NewNotifier() Notifier {
	return Notifier{
		hUp:   make(chan mgm.Host, 32),
		hDel:  make(chan int64, 32),
		hStat: make(chan mgm.HostStat, 32),
		rUp:   make(chan mgm.Region, 32),
		rStat: make(chan mgm.RegionStat, 32),
		rDel:  make(chan mgm.Region, 32),
		eUp:   make(chan mgm.Estate, 32),
		eDel:  make(chan mgm.Estate, 32),
		jUp:   make(chan mgm.Job, 32),
		jDel:  make(chan mgm.Job, 32),
	}
}

//HostUpdated notifies that a host has been updated
func (n Notifier) HostUpdated(h mgm.Host) {
	n.hUp <- h
}

// HostRemoved notifies that a host has been deleted
func (n Notifier) HostRemoved(h int64) {
	n.hDel <- h
}

// HostStat notifies that a host status has updated
func (n Notifier) HostStat(h mgm.HostStat) {
	n.hStat <- h
}

//RegionUpdated notifies that a region has been updated
func (n Notifier) RegionUpdated(r mgm.Region) {
	n.rUp <- r
}

//RegionDeleted notifies that a region has been deleted
func (n Notifier) RegionDeleted(r mgm.Region) {
	n.rDel <- r
}

//RegionStat notifies that a regions status has been updated
func (n Notifier) RegionStat(s mgm.RegionStat) {
	n.rStat <- s
}

//EstateUpdated notifies that an estate has been modified
func (n Notifier) EstateUpdated(e mgm.Estate) {
	n.eUp <- e
}

//EstateDeleted notifies that an estate has been deleted
func (n Notifier) EstateDeleted(e mgm.Estate) {
	n.eDel <- e
}

//JobUpdated notifies that a job record has been created/updated
func (n Notifier) JobUpdated(j mgm.Job) {
	n.jUp <- j
}

//JobDeleted notifies that a job record has been removed
func (n Notifier) JobDeleted(j mgm.Job) {
	n.jDel <- j
}
