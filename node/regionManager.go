package main

type regionManager interface {
	PurgeRegionDirectory()
}

func newRegionManager(binDir string, regionDir string) regionManager {
	return regMgr{binDir, regionDir}
}

type regMgr struct {
	copyFrom  string
	regionDir string
}

func (rm regMgr) PurgeRegionDirectory() {

}
