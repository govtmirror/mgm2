package remote

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/m-o-s-e-s/mgm/mgm"
)

func (r region) WriteRegionINI(reg mgm.Region) error {
	regionsINI := path.Join(r.dir, "Regions", "Regions.ini")

	content :=
		`[%s]
RegionUUID = %s
Location = %d,%d
InternalAddress = 0.0.0.0
InternalPort = %d
SizeX = %d
SizeY = %d
AllowAlternatePorts = False
ExternalHostName = %s
`

	content = fmt.Sprintf(
		content,
		reg.Name,
		reg.UUID,
		reg.LocX,
		reg.LocY,
		reg.HTTPPort,
		reg.Size*256,
		reg.Size*256,
		r.hostName,
	)

	err := ioutil.WriteFile(regionsINI, []byte(content), 0644)
	return err
}

func (r region) WriteOpensimINI(configs []mgm.ConfigOption) error {
	opensimINI := path.Join(r.dir, "OpenSim.ini")

	cfgs := make(map[string]map[string]string)
	for _, cfg := range configs {
		if _, ok := cfgs[cfg.Section]; !ok {
			cfgs[cfg.Section] = make(map[string]string)
		}
		cfgs[cfg.Section][cfg.Item] = cfg.Content
	}

	//write the configuration into a buffer
	var buffer bytes.Buffer

	for section, m := range cfgs {
		buffer.WriteString(fmt.Sprintf("[%s]\n", section))
		for item, content := range m {
			buffer.WriteString(fmt.Sprintf("  %s = \"%s\"\n", item, content))
		}
	}

	f, err := os.Create(opensimINI)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = buffer.WriteTo(f)

	return err
}
