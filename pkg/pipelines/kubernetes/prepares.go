package kubernetes

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
)

type NoClusterInfo struct {
	common.KubePrepare
}

func (n *NoClusterInfo) PreCheck(_ connector.Runtime) (bool, error) {
	if v, ok := n.RootCache.Get(ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)
		if cluster.Info == "" {
			return true, nil
		}
	} else {
		return false, errors.New("get kubernetes cluster status by pipeline cache failed")
	}
	return false, nil
}

type NodeInCluster struct {
	common.KubePrepare
	Not bool
}

func (n *NodeInCluster) PreCheck(runtime connector.Runtime) (bool, error) {
	host := runtime.RemoteHost()
	if v, ok := n.RootCache.Get(ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)
		var versionOk bool
		if res, ok := cluster.NodesInfo[host.GetName()]; ok && res != "" {
			versionOk = true
		}
		_, ipOk := cluster.NodesInfo[host.GetInternalAddress()]
		if n.Not {
			return !(versionOk || ipOk), nil
		}
		return versionOk || ipOk, nil
	} else {
		return false, errors.New("get kubernetes cluster status by pipeline cache failed")
	}
}

type ClusterIsExist struct {
	common.KubePrepare
	Not bool
}

func (c *ClusterIsExist) PreCheck(_ connector.Runtime) (bool, error) {
	if exist, ok := c.RootCache.GetMustBool(ClusterExist); ok {
		if c.Not {
			return !exist, nil
		}
		return exist, nil
	} else {
		return false, errors.New("get kubernetes cluster status by pipeline cache failed")
	}
}
