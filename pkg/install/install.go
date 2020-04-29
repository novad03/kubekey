package install

import (
	"fmt"
	"github.com/pixiake/kubekey/pkg/cluster/etcd"
	"github.com/pixiake/kubekey/pkg/cluster/kubernetes"
	"github.com/pixiake/kubekey/pkg/cluster/preinstall"
	"github.com/pixiake/kubekey/pkg/config"
	"github.com/pixiake/kubekey/pkg/container-engine/docker"
	"github.com/pixiake/kubekey/pkg/images"
	"github.com/pixiake/kubekey/pkg/plugins/network"
	"github.com/pixiake/kubekey/pkg/plugins/storage"
	"github.com/pixiake/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func CreateCluster(clusterCfgFile string, logger *log.Logger, addons, pkg string, verbose bool) error {
	cfg, err := config.ParseClusterCfg(clusterCfgFile, logger)
	if err != nil {
		return errors.Wrap(err, "failed to download cluster config")
	}

	//out, _ := json.MarshalIndent(cfg, "", "  ")
	//fmt.Println(string(out))
	if err := preinstall.Prepare(&cfg.Spec, logger); err != nil {
		return errors.Wrap(err, "failed to load kube binarys")
	}
	return NewExecutor(&cfg.Spec, logger, verbose).Execute()
}

func ExecTasks(mgr *manager.Manager) error {
	createTasks := []manager.Task{
		{Task: preinstall.InitOS, ErrMsg: "failed to download kube binaries"},
		{Task: docker.InstallerDocker, ErrMsg: "failed to install docker"},
		{Task: kubernetes.SyncKubeBinaries, ErrMsg: "failed to sync kube binaries"},
		{Task: images.PreDownloadImages, ErrMsg: "failed to pre-download images"},
		{Task: etcd.GenerateEtcdCerts, ErrMsg: "failed to generate etcd certs"},
		{Task: etcd.SyncEtcdCertsToMaster, ErrMsg: "failed to sync etcd certs"},
		{Task: etcd.GenerateEtcdService, ErrMsg: "failed to start etcd cluster"},
		{Task: kubernetes.ConfigureKubeletService, ErrMsg: "failed to sync kube binaries"},
		{Task: kubernetes.InitKubernetesCluster, ErrMsg: "failed to init kubernetes cluster"},
		{Task: network.DeployNetworkPlugin, ErrMsg: "failed to deploy network plugin"},
		{Task: kubernetes.GetJoinNodesCmd, ErrMsg: "failed to get join cmd"},
		{Task: kubernetes.JoinNodesToCluster, ErrMsg: "failed to join node"},
		{Task: storage.DeployStoragePlugins, ErrMsg: "failed to deploy storage plugin"},
	}

	for _, step := range createTasks {
		if err := step.Run(mgr); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}

	fmt.Printf("\n\033[1;36;40m%s\033[0m\n", "Successful.")
	return nil
}
