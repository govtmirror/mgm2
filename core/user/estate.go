package user

import "github.com/m-o-s-e-s/mgm/mgm"

// GetEstates gets an array of all current estates
func (m Manager) GetEstates() []mgm.Estate {
	m.estateMutex.Lock()
	defer m.estateMutex.Unlock()
	t := []mgm.Estate{}
	for _, e := range m.estates {
		t = append(t, e)
	}
	return t
}
