package session

import "github.com/m-o-s-e-s/mgm/mgm"

//Notifier hold a bunch of channels for communicating with user sessions
type Notifier struct {
	hUp   chan mgm.Host
	hDel  chan mgm.Host
	hStat chan mgm.HostStat
	rUp   chan mgm.Region
	rDel  chan mgm.Region
	rStat chan mgm.RegionStat
	eUp   chan mgm.Estate
	eDel  chan mgm.Estate
}

//NewNotifier constructs a Notifier, initializing all internal data structures
func NewNotifier() Notifier {
	return Notifier{
		hUp:   make(chan mgm.Host, 32),
		hDel:  make(chan mgm.Host, 32),
		hStat: make(chan mgm.HostStat, 32),
		rUp:   make(chan mgm.Region, 32),
		rStat: make(chan mgm.RegionStat, 32),
		rDel:  make(chan mgm.Region, 32),
		eUp:   make(chan mgm.Estate, 32),
		eDel:  make(chan mgm.Estate, 32),
	}
}

//HostUpdated notifies that a host has been added/updated
func (n Notifier) HostUpdated(h mgm.Host) {
	n.hUp <- h
}

// HostDeleted notifies that a host has been deleted
func (n Notifier) HostDeleted(h mgm.Host) {
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
