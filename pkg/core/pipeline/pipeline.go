package pipeline

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/pkg/errors"
	"sync"
)

var logo = `

 _   __      _          _   __           
| | / /     | |        | | / /           
| |/ / _   _| |__   ___| |/ /  ___ _   _ 
|    \| | | | '_ \ / _ \    \ / _ \ | | |
| |\  \ |_| | |_) |  __/ |\  \  __/ |_| |
\_| \_/\__,_|_.__/ \___\_| \_/\___|\__, |
                                    __/ |
                                   |___/

`

type Pipeline struct {
	Name            string
	Modules         []modules.Module
	Runtime         connector.Runtime
	PipelineCache   *cache.Cache
	ModuleCachePool sync.Pool
}

func (p *Pipeline) Init() error {
	fmt.Print(logo)
	p.PipelineCache = cache.NewCache()
	if err := p.Runtime.GenerateWorkDir(); err != nil {
		return err
	}
	if err := p.Runtime.InitLogger(); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) Start() error {
	if err := p.Init(); err != nil {
		return errors.Wrapf(err, "Pipeline[%s] exec failed", p.Name)
	}
	for i := range p.Modules {
		m := p.Modules[i]
		if m.IsSkip() {
			continue
		}
		if err := p.RunModule(m); err != nil {
			return errors.Wrapf(err, "Pipeline[%s] exec failed", p.Name)
		}
	}
	logger.Log.Infof("Pipeline[%s] execute successful", p.Name)
	return nil
}

func (p *Pipeline) RunModule(m modules.Module) error {
	moduleCache := p.newModuleCache()
	defer p.releaseModuleCache(moduleCache)
	m.Default(p.Runtime, p.PipelineCache, moduleCache)
	m.AutoAssert()
	m.Init()
	m.Slogan()
	switch m.Is() {
	case modules.TaskModuleType:
		if err := m.Run(); err != nil {
			return err
		}
	case modules.ServerModuleType:
		go m.Run()
	default:
		if err := m.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) newModuleCache() *cache.Cache {
	moduleCache, ok := p.ModuleCachePool.Get().(*cache.Cache)
	if ok {
		return moduleCache
	}
	return cache.NewCache()
}

func (p *Pipeline) releaseModuleCache(c *cache.Cache) {
	c.Clean()
	p.ModuleCachePool.Put(c)
}
