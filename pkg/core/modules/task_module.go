package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/pkg/errors"
)

type BaseTaskModule struct {
	BaseModule
	Tasks []Task
}

func (b *BaseTaskModule) Default(runtime connector.Runtime, rootCache *cache.Cache, moduleCache *cache.Cache) {
	if b.Name == "" {
		b.Name = DefaultTaskModuleName
	}

	b.Runtime = runtime
	b.RootCache = rootCache
	b.Cache = moduleCache
}

func (b *BaseTaskModule) Init() {
}

func (b *BaseTaskModule) Is() string {
	return TaskModuleType
}

func (b *BaseTaskModule) Run() error {
	for i := range b.Tasks {
		task := b.Tasks[i]
		task.Init(b.Name, b.Runtime, b.Cache, b.RootCache)
		if res := task.Execute(); res.IsFailed() {
			return errors.Wrapf(res.CombineErr(), "Module[%s] exec failed", b.Name)
		}
	}
	return nil
}
