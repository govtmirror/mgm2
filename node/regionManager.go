package node

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/M-O-S-E-S/mgm/core"
	"github.com/M-O-S-E-S/mgm/mgm"
)

// RegionManager interfaces with the region management objects
type RegionManager interface {
	Initialize() error
	AddRegion(mgm.Region) error
	RemoveRegion(mgm.Region) error
}

// NewRegionManager constructs a region manager for use
func NewRegionManager(binDir string, regionDir string, log core.Logger) RegionManager {
	return regMgr{
		copyFrom:  binDir,
		regionDir: regionDir,
		logger:    log,
	}
}

type regMgr struct {
	copyFrom  string
	regionDir string
	logger    core.Logger
	regions   []mgm.Region
}

func (rm regMgr) AddRegion(r mgm.Region) error {
	return nil
}

func (rm regMgr) RemoveRegion(r mgm.Region) error {
	return nil
}

func (rm regMgr) Initialize() error {
	//confirm binaries are present
	if _, err := os.Stat(path.Join(rm.copyFrom, "OpenSim.exe")); os.IsNotExist(err) {
		return errors.New("Opensim source directory does not exist")
	}
	//confirm regions directory exists
	if _, err := os.Stat(rm.regionDir); os.IsNotExist(err) {
		return errors.New("Regions directory does not exist")
	}

	files, err := ioutil.ReadDir(rm.regionDir)
	if err != nil {
		return err
	}
	rm.logger.Info("Purging %v old region record(s)", len(files))
	for _, f := range files {
		err = os.RemoveAll(path.Join(rm.regionDir, f.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}
