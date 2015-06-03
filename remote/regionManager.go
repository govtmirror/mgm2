package remote

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// RegionManager interfaces with the region management objects
type RegionManager interface {
	Initialize() error
	AddRegion(mgm.Region) (Region, error)
	RemoveRegion(mgm.Region) error
}

// NewRegionManager constructs a region manager for use
func NewRegionManager(binDir string, regionDir string, hostname string, log logger.Log) RegionManager {
	return regMgr{
		copyFrom:  binDir,
		regionDir: regionDir,
		hostName:  hostname,
		logger:    logger.Wrap("RegionManager", log),
	}
}

type regMgr struct {
	copyFrom  string
	regionDir string
	logger    logger.Log
	hostName  string
	regions   []mgm.Region
}

func (rm regMgr) AddRegion(r mgm.Region) (Region, error) {
	path, err := rm.copyBinaries(r.UUID.String())
	if err != nil {
		return region{}, err
	}
	reg := NewRegion(r, path, rm.hostName, rm.logger)

	return reg, nil
}

func (rm regMgr) RemoveRegion(r mgm.Region) error {
	return nil
}

func (rm regMgr) copyBinaries(name string) (string, error) {
	copyTo := filepath.Join(rm.regionDir, name)
	err := os.Mkdir(copyTo, 0700)
	if err != nil {
		return "", err
	}
	err = filepath.Walk(rm.copyFrom, func(path string, info os.FileInfo, err error) error {
		if path == rm.copyFrom {
			return nil
		}

		if info.IsDir() {
			return os.Mkdir(strings.Replace(path, rm.copyFrom, copyTo, 1), 0700)
		}
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(strings.Replace(path, rm.copyFrom, copyTo, 1))
		if err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			return err
		}
		return dst.Close()
	})
	return copyTo, err
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
