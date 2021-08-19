package prepare

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
)

type Prepare interface {
	PreCheck() (bool, error)
	Init(mgr *config.Runtime, cache *cache.Cache, rootCache *cache.Cache)
}
