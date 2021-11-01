package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
)

type Module interface {
	IsSkip() bool
	Default(runtime connector.Runtime, pipelineCache *cache.Cache, moduleCache *cache.Cache)
	Init()
	Is() string
	Run() error
	Slogan()
	AutoAssert()
}

type Task interface {
	Init(moduleName string, runtime connector.Runtime, moduleCache *cache.Cache, pipelineCache *cache.Cache)
	Execute() *ending.TaskResult
}
