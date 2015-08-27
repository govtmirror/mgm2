package user

import "github.com/m-o-s-e-s/mgm/mgm"

// GetGroups gets an array of all current estates
func (m Manager) GetGroups() []mgm.Group {
	m.groupMutex.Lock()
	defer m.groupMutex.Unlock()
	t := []mgm.Group{}
	for _, g := range m.groups {
		t = append(t, g)
	}
	return t
}
