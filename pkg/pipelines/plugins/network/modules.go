package network

import (
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/images"
	"github.com/kubesphere/kubekey/pkg/pipelines/plugins/network/templates"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"path/filepath"
)

type DeployNetworkPluginModule struct {
	common.KubeModule
}

func (d *DeployNetworkPluginModule) Init() {
	d.Name = "DeployNetworkPluginModule"

	switch d.KubeConf.Cluster.Network.Plugin {
	case common.Calico:
		d.Tasks = deployCalico(d)
	case common.Flannel:
		d.Tasks = deployFlannel(d)
	case common.Cilium:
		d.Tasks = deployCilium(d)
	case common.Kubeovn:
	default:
		return
	}
}

func deployCalico(d *DeployNetworkPluginModule) []modules.Task {
	generateCalicoOld := &modules.RemoteTask{
		Name:  "GenerateCalico",
		Desc:  "generate calico",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(OldK8sVersion),
		},
		Action: &action.Template{
			Template: templates.CalicoOld,
			Dst:      filepath.Join(common.KubeConfigDir, templates.CalicoOld.Name()),
			Data: util.Data{
				"KubePodsCIDR":           d.KubeConf.Cluster.Network.KubePodsCIDR,
				"CalicoCniImage":         images.GetImage(d.Runtime, d.KubeConf, "calico-cni").ImageName(),
				"CalicoNodeImage":        images.GetImage(d.Runtime, d.KubeConf, "calico-node").ImageName(),
				"CalicoFlexvolImage":     images.GetImage(d.Runtime, d.KubeConf, "calico-flexvol").ImageName(),
				"CalicoControllersImage": images.GetImage(d.Runtime, d.KubeConf, "calico-kube-controllers").ImageName(),
				"TyphaEnabled":           len(d.Runtime.GetHostsByRole(common.K8s)) > 50,
				"VethMTU":                d.KubeConf.Cluster.Network.Calico.VethMTU,
				"NodeCidrMaskSize":       d.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
				"IPIPMode":               d.KubeConf.Cluster.Network.Calico.IPIPMode,
				"VXLANMode":              d.KubeConf.Cluster.Network.Calico.VXLANMode,
			},
		},
		Parallel: true,
	}

	generateCalicoNew := &modules.RemoteTask{
		Name:  "GenerateCalico",
		Desc:  "generate calico",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&OldK8sVersion{Not: true},
		},
		Action: &action.Template{
			Template: templates.CalicoNew,
			Dst:      filepath.Join(common.KubeConfigDir, templates.CalicoNew.Name()),
			Data: util.Data{
				"KubePodsCIDR":            d.KubeConf.Cluster.Network.KubePodsCIDR,
				"CalicoCniImage":          images.GetImage(d.Runtime, d.KubeConf, "calico-cni").ImageName(),
				"CalicoNodeImage":         images.GetImage(d.Runtime, d.KubeConf, "calico-node").ImageName(),
				"CalicoFlexvolImage":      images.GetImage(d.Runtime, d.KubeConf, "calico-flexvol").ImageName(),
				"CalicoControllersImage":  images.GetImage(d.Runtime, d.KubeConf, "calico-kube-controllers").ImageName(),
				"CalicoTyphaImage":        images.GetImage(d.Runtime, d.KubeConf, "calico-typha").ImageName(),
				"TyphaEnabled":            len(d.Runtime.GetHostsByRole(common.K8s)) > 50,
				"VethMTU":                 d.KubeConf.Cluster.Network.Calico.VethMTU,
				"NodeCidrMaskSize":        d.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
				"IPIPMode":                d.KubeConf.Cluster.Network.Calico.IPIPMode,
				"VXLANMode":               d.KubeConf.Cluster.Network.Calico.VXLANMode,
				"ConatinerManagerIsIsula": d.KubeConf.Cluster.Kubernetes.ContainerManager == "isula",
			},
		},
		Parallel: true,
	}

	deploy := &modules.RemoteTask{
		Name:     "DeployCalico",
		Desc:     "deploy calico",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(DeployNetworkPlugin),
		Parallel: true,
		Retry:    5,
	}

	if CompareVersionLater(d.KubeConf.Cluster.Kubernetes.Version, "v1.16.0") {
		return []modules.Task{
			generateCalicoNew,
			deploy,
		}
	} else {
		return []modules.Task{
			generateCalicoOld,
			deploy,
		}
	}
}

func deployFlannel(d *DeployNetworkPluginModule) []modules.Task {
	generateFlannel := &modules.RemoteTask{
		Name:  "GenerateFlannel",
		Desc:  "generate flannel",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(OldK8sVersion),
		},
		Action: &action.Template{
			Template: templates.Flannel,
			Dst:      filepath.Join(common.KubeConfigDir, templates.Flannel.Name()),
			Data: util.Data{
				"KubePodsCIDR": d.KubeConf.Cluster.Network.KubePodsCIDR,
				"FlannelImage": images.GetImage(d.Runtime, d.KubeConf, "flannel").ImageName(),
				"BackendMode":  d.KubeConf.Cluster.Network.Flannel.BackendMode,
			},
		},
		Parallel: true,
	}

	deploy := &modules.RemoteTask{
		Name:     "DeployFlannel",
		Desc:     "deploy flannel",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(DeployNetworkPlugin),
		Parallel: true,
		Retry:    5,
	}

	return []modules.Task{
		generateFlannel,
		deploy,
	}
}

func deployCilium(d *DeployNetworkPluginModule) []modules.Task {
	generateCilium := &modules.RemoteTask{
		Name:  "GenerateCilium",
		Desc:  "generate cilium",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(OldK8sVersion),
		},
		Action: &action.Template{
			Template: templates.Cilium,
			Dst:      filepath.Join(common.KubeConfigDir, templates.Cilium.Name()),
			Data: util.Data{
				"KubePodsCIDR":         d.KubeConf.Cluster.Network.KubePodsCIDR,
				"NodeCidrMaskSize":     d.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
				"CiliumImage":          images.GetImage(d.Runtime, d.KubeConf, "cilium").ImageName(),
				"OperatorGenericImage": images.GetImage(d.Runtime, d.KubeConf, "operator-generic").ImageName(),
			},
		},
		Parallel: true,
	}

	deploy := &modules.RemoteTask{
		Name:     "DeployCilium",
		Desc:     "deploy cilium",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(DeployNetworkPlugin),
		Parallel: true,
		Retry:    5,
	}

	return []modules.Task{
		generateCilium,
		deploy,
	}
}

func CompareVersionLater(version string, compare string) bool {
	cmp, err := versionutil.MustParseSemantic(version).Compare(compare)
	if err != nil {
		logger.Log.Fatal("unknown kubernetes version")
	}
	// old version
	if cmp == -1 {
		return false
	} else {
		// new version
		return true
	}
}
