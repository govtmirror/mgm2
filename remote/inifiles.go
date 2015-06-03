package remote

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/m-o-s-e-s/mgm/mgm"
)

func (r region) WriteRegionINI(reg mgm.Region) error {
	regionsINI := path.Join(r.dir, "Regions", "Regions.ini")

	content :=
		`[%s]
  RegionUUID = "%s"
  Location = "%d,%d"
  InternalAddress = "0.0.0.0"
  InternalPort = %d
  SizeX = %d
  SizeY = %d
  AllowAlternatePorts = False
  ExternalHostName = "%s"`

	content = fmt.Sprintf(
		content,
		reg.Name,
		reg.UUID,
		reg.LocX,
		reg.LocY,
		r.hostName,
	)

	err := ioutil.WriteFile(regionsINI, []byte(content), 0644)
	return err
}

func (r region) WriteOpensimINI(defaults []mgm.ConfigOption, config []mgm.ConfigOption) error {
	return nil
}
