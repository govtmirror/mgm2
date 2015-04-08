package mgm

type regionManager struct {
  
}

func (rm * regionManager) newRegion() (r * region){
  return &region{}
}