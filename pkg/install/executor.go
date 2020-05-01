package install

import (
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	cluster *kubekeyapi.K2ClusterSpec
	logger  *log.Logger
	Verbose bool
}

func NewExecutor(cluster *kubekeyapi.K2ClusterSpec, logger *log.Logger, verbose bool) *Executor {
	return &Executor{
		cluster: cluster,
		logger:  logger,
		Verbose: verbose,
	}
}

func (executor *Executor) Execute() error {
	mgr, err := executor.createManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
}

func (executor *Executor) createManager() (*manager.Manager, error) {
	mgr := &manager.Manager{}
	allNodes, etcdNodes, masterNodes, workerNodes, k8sNodes, clientNode := executor.cluster.GroupHosts()
	mgr.AllNodes = allNodes
	mgr.EtcdNodes = etcdNodes
	mgr.MasterNodes = masterNodes
	mgr.WorkerNodes = workerNodes
	mgr.K8sNodes = k8sNodes
	mgr.ClientNode = clientNode
	mgr.Cluster = executor.cluster
	mgr.Connector = ssh.NewConnector()
	mgr.Logger = executor.logger
	mgr.Verbose = executor.Verbose

	return mgr, nil
}
