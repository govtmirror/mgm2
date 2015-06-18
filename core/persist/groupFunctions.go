package persist

import "github.com/m-o-s-e-s/mgm/mgm"

func (m mgmDB) GetGroups() []mgm.Group {
	var groups []mgm.Group
	r := mgmReq{}
	r.request = "GetGroups"
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return groups
		}
		groups = append(groups, h.(mgm.Group))
	}
}
