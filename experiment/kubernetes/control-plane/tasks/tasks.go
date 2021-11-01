package tasks

import (
	"github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/pipline"
	"github.com/kubesphere/kubekey/pkg/util/manager"
)

var (
	initClusterCmd   = "kubeadm init -f kubeadm.conf"
	getKubeConfigCmd = "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	addNodeCmdTmpl   = "kubeadm join {{ .ApiServer }} --token {{ .Token }} --discovery-token-ca-cert-hash {{ .Hash }}"

	mgr         = manager.Manager{}
	InitCluster = pipline.Task{
		Name:  "Init Cluster",
		Hosts: []v1alpha1.HostCfg{mgr.MasterNodes[0]},
		Action: &pipline.Command{
			Cmd: initClusterCmd,
		},
		Env: nil,
		Vars: pipline.Vars{
			"kubernetes": config.GetConfig().Cluster.Kubernetes.ClusterName,
		},
		Parallel:    false,
		Prepare:     &pipline.Condition{Cond: true},
		IgnoreError: false,
	}

	GetKubeConfig = pipline.Task{
		Name:    "Get KubeConfig",
		Hosts:   mgr.MasterNodes,
		Action:  &pipline.Command{Cmd: getKubeConfigCmd},
		Prepare: &pipline.Condition{Cond: true},
	}
)
