package k3s

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/k3s/templates"
	"path/filepath"
)

type StatusModule struct {
	common.KubeModule
}

func (s *StatusModule) Init() {
	s.Name = "StatusModule"

	cluster := NewK3sStatus()
	s.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &module.RemoteTask{
		Name:     "GetClusterStatus",
		Desc:     "Get k3s cluster status",
		Hosts:    s.Runtime.GetHostsByRole(common.Master),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	s.Tasks = []module.Task{
		clusterStatus,
	}
}

type InstallKubeBinariesModule struct {
	common.KubeModule
}

func (i *InstallKubeBinariesModule) Init() {
	i.Name = "InstallKubeBinariesModule"

	syncBinary := &module.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize k3s binaries",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	killAllScript := &module.RemoteTask{
		Name:    "GenerateK3sKillAllScript",
		Desc:    "Generate k3s killall.sh script",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &NodeInCluster{Not: true},
		Action: &action.Template{
			Template: templates.K3sKillallScript,
			Dst:      filepath.Join("/usr/local/bin", templates.K3sKillallScript.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	uninstallScript := &module.RemoteTask{
		Name:    "GenerateK3sUninstallScript",
		Desc:    "Generate k3s uninstall script",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &NodeInCluster{Not: true},
		Action: &action.Template{
			Template: templates.K3sUninstallScript,
			Dst:      filepath.Join("/usr/local/bin", templates.K3sUninstallScript.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	chmod := &module.RemoteTask{
		Name:     "ChmodScript",
		Desc:     "Chmod +x k3s script ",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(ChmodScript),
		Parallel: true,
		Retry:    2,
	}

	i.Tasks = []module.Task{
		syncBinary,
		killAllScript,
		uninstallScript,
		chmod,
	}
}

type InitClusterModule struct {
	common.KubeModule
}

func (i *InitClusterModule) Init() {
	i.Name = "K3sInitClusterModule"

	k3sService := &module.RemoteTask{
		Name:  "GenerateK3sService",
		Desc:  "Generate k3s Service",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(GenerateK3sService),
		Parallel: true,
	}

	k3sEnv := &module.RemoteTask{
		Name:  "GenerateK3sServiceEnv",
		Desc:  "Generate k3s service env",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(GenerateK3sServiceEnv),
		Parallel: true,
	}

	enableK3s := &module.RemoteTask{
		Name:  "EnableK3sService",
		Desc:  "Enable k3s service",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(EnableK3sService),
		Parallel: true,
	}

	copyKubeConfig := &module.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy k3s.yaml to ~/.kube/config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(CopyK3sKubeConfig),
		Parallel: true,
	}

	addMasterTaint := &module.RemoteTask{
		Name:  "AddMasterTaint",
		Desc:  "Add master taint",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			&common.IsWorker{Not: true},
		},
		Action:   new(AddMasterTaint),
		Parallel: true,
		Retry:    5,
	}

	addWorkerLabel := &module.RemoteTask{
		Name:  "AddWorkerLabel",
		Desc:  "Add worker label",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	i.Tasks = []module.Task{
		k3sService,
		k3sEnv,
		enableK3s,
		copyKubeConfig,
		addMasterTaint,
		addWorkerLabel,
	}
}

type JoinNodesModule struct {
	common.KubeModule
}

func (j *JoinNodesModule) Init() {
	j.Name = "K3sJoinNodesModule"

	k3sService := &module.RemoteTask{
		Name:  "GenerateK3sService",
		Desc:  "Generate k3s Service",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(GenerateK3sService),
		Parallel: true,
	}

	k3sEnv := &module.RemoteTask{
		Name:  "GenerateK3sServiceEnv",
		Desc:  "Generate k3s service env",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(GenerateK3sServiceEnv),
		Parallel: true,
	}

	enableK3s := &module.RemoteTask{
		Name:  "EnableK3sService",
		Desc:  "Enable k3s service",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(EnableK3sService),
		Parallel: true,
	}

	copyKubeConfigForMaster := &module.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy k3s.yaml to ~/.kube/config",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(CopyK3sKubeConfig),
		Parallel: true,
	}

	syncKubeConfigToWorker := &module.RemoteTask{
		Name:  "SyncKubeConfigToWorker",
		Desc:  "Synchronize kube config to worker",
		Hosts: j.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.OnlyWorker),
		},
		Action:   new(SyncKubeConfigToWorker),
		Parallel: true,
	}

	addMasterTaint := &module.RemoteTask{
		Name:  "AddMasterTaint",
		Desc:  "Add master taint",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			&common.IsWorker{Not: true},
		},
		Action:   new(AddMasterTaint),
		Parallel: true,
		Retry:    5,
	}

	addWorkerLabel := &module.RemoteTask{
		Name:  "AddWorkerLabel",
		Desc:  "Add worker label",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	j.Tasks = []module.Task{
		k3sService,
		k3sEnv,
		enableK3s,
		copyKubeConfigForMaster,
		syncKubeConfigToWorker,
		addMasterTaint,
		addWorkerLabel,
	}
}

type DeleteClusterModule struct {
	common.KubeModule
}

func (d *DeleteClusterModule) Init() {
	d.Name = "DeleteClusterModule"

	execScript := &module.RemoteTask{
		Name:     "ExecUninstallScript",
		Desc:     "Exec k3s uninstall script",
		Hosts:    d.Runtime.GetHostsByRole(common.K8s),
		Action:   new(ExecUninstallScript),
		Parallel: true,
	}

	d.Tasks = []module.Task{
		execScript,
	}
}
