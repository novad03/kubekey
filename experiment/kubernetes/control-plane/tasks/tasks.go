package tasks

import (
	"github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/action"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
)

var (
	initClusterCmd   = "kubeadm init -f kubeadm.conf"
	getKubeConfigCmd = "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	addNodeCmdTmpl   = "kubeadm join {{ .ApiServer }} --token {{ .Token }} --discovery-token-ca-cert-hash {{ .Hash }}"

	mgr         = config.GetManager()
	host        = mgr.Runner.Host
	InitCluster = pipeline.Task{
		Name:  "Init Cluster",
		Hosts: []v1alpha1.HostCfg{mgr.MasterNodes[0]},
		Action: &action.Command{
			Cmd: initClusterCmd,
		},
		Env: nil,
		Vars: pipeline.Vars{
			"kubernetes": config.GetManager().Cluster.Kubernetes.ClusterName,
			"ipaddr":     host.InternalAddress,
		},
		Parallel:    false,
		Prepare:     &pipeline.Condition{Cond: true},
		IgnoreError: false,
	}

	GetKubeConfig = pipeline.Task{
		Name:    "Get KubeConfig",
		Hosts:   mgr.MasterNodes,
		Action:  &action.Command{Cmd: getKubeConfigCmd},
		Prepare: &pipeline.Condition{Cond: true},
		// todo 放指针
	}
)
