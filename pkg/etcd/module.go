package etcd

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/etcd/templates"
	"path/filepath"
)

type PreCheckModule struct {
	common.KubeModule
}

func (p *PreCheckModule) Init() {
	p.Name = "ETCDPreCheckModule"
	p.Desc = "Get ETCD cluster status"

	getStatus := &module.RemoteTask{
		Name:     "GetETCDStatus",
		Desc:     "Get etcd status",
		Hosts:    p.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(GetStatus),
		Parallel: false,
		Retry:    0,
	}
	p.Tasks = []module.Task{
		getStatus,
	}
}

type CertsModule struct {
	common.KubeModule
}

func (c *CertsModule) Init() {
	c.Name = "CertsModule"
	c.Desc = "Sign ETCD cluster certs"

	generateCertsScript := &module.RemoteTask{
		Name:    "GenerateCertsScript",
		Desc:    "Generate certs script",
		Hosts:   c.Runtime.GetHostsByRole(common.ETCD),
		Prepare: new(FirstETCDNode),
		Action: &action.Template{
			Template: templates.EtcdSslScript,
			Dst:      filepath.Join(common.ETCDCertDir, templates.EtcdSslScript.Name()),
			Data: util.Data{
				"Masters": templates.GenerateHosts(c.Runtime.GetHostsByRole(common.ETCD)),
				"Hosts":   templates.GenerateHosts(c.Runtime.GetHostsByRole(common.Master)),
			},
		},
		Parallel: true,
		Retry:    1,
	}

	dnsList, ipList := templates.DNSAndIp(c.KubeConf)
	generateOpenSSLConf := &module.RemoteTask{
		Name:    "GenerateOpenSSLConf",
		Desc:    "Generate OpenSSL config",
		Hosts:   c.Runtime.GetHostsByRole(common.ETCD),
		Prepare: new(FirstETCDNode),
		Action: &action.Template{
			Template: templates.ETCDOpenSSLConf,
			Dst:      filepath.Join(common.ETCDCertDir, templates.ETCDOpenSSLConf.Name()),
			Data: util.Data{
				"Dns": dnsList,
				"Ips": ipList,
			},
		},
		Parallel: true,
		Retry:    1,
	}

	execCertsScript := &module.RemoteTask{
		Name:     "ExecCertsScript",
		Desc:     "Exec certs script",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(ExecCertsScript),
		Parallel: true,
		Retry:    1,
	}

	syncCertsFile := &module.RemoteTask{
		Name:     "SyncCertsFile",
		Desc:     "Synchronize certs file",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &FirstETCDNode{Not: true},
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	syncCertsToMaster := &module.RemoteTask{
		Name:     "SyncCertsFileToMaster",
		Desc:     "Synchronize certs file to master",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  &common.OnlyETCD{Not: true},
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	c.Tasks = []module.Task{
		generateCertsScript,
		generateOpenSSLConf,
		execCertsScript,
		syncCertsFile,
		syncCertsToMaster,
	}
}

type InstallETCDBinaryModule struct {
	common.KubeModule
}

func (i *InstallETCDBinaryModule) Init() {
	i.Name = "InstallETCDBinaryModule"
	i.Desc = "Install ETCD cluster"

	installETCDBinary := &module.RemoteTask{
		Name:     "InstallETCDBinary",
		Desc:     "Install etcd using binary",
		Hosts:    i.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(InstallETCDBinary),
		Parallel: true,
		Retry:    1,
	}

	generateETCDService := &module.RemoteTask{
		Name:  "GenerateETCDService",
		Desc:  "Generate etcd service",
		Hosts: i.Runtime.GetHostsByRole(common.ETCD),
		Action: &action.Template{
			Template: templates.ETCDService,
			Dst:      "/etc/systemd/system/etcd.service",
		},
		Parallel: true,
		Retry:    1,
	}

	accessAddress := &module.RemoteTask{
		Name:     "GenerateAccessAddress",
		Desc:     "Generate access address",
		Hosts:    i.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(GenerateAccessAddress),
		Parallel: true,
		Retry:    1,
	}

	i.Tasks = []module.Task{
		installETCDBinary,
		generateETCDService,
		accessAddress,
	}
}

type ConfigureModule struct {
	common.KubeModule
}

func (e *ConfigureModule) Init() {
	e.Name = "ETCDConfigureModule"
	e.Desc = "Configure ETCD cluster"

	if v, ok := e.PipelineCache.Get(common.ETCDCluster); ok {
		cluster := v.(*EtcdCluster)
		if !cluster.clusterExist {
			e.Tasks = handleNewCluster(e)
		} else {
			e.Tasks = handleExistCluster(e)
		}
	}
}

func handleNewCluster(c *ConfigureModule) []module.Task {

	existETCDHealthCheck := &module.RemoteTask{
		Name:     "ExistETCDHealthCheck",
		Desc:     "Health check on exist etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(NodeETCDExist),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	generateETCDConfig := &module.RemoteTask{
		Name:     "GenerateETCDConfig",
		Desc:     "Generate etcd.env config on new etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(GenerateConfig),
		Parallel: false,
	}

	allRefreshETCDConfig := &module.RemoteTask{
		Name:     "AllRefreshETCDConfig",
		Desc:     "Refresh etcd.env config on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(RefreshConfig),
		Parallel: false,
	}

	restart := &module.RemoteTask{
		Name:     "RestartETCD",
		Desc:     "Restart etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(RestartETCD),
		Parallel: true,
	}

	allETCDNodeHealthCheck := &module.RemoteTask{
		Name:     "AllETCDNodeHealthCheck",
		Desc:     "Health check on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	refreshETCDConfigToExist := &module.RemoteTask{
		Name:     "RefreshETCDConfigToExist",
		Desc:     "Refresh etcd.env config to exist mode on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   &RefreshConfig{ToExisting: true},
		Parallel: false,
	}

	tasks := []module.Task{
		existETCDHealthCheck,
		generateETCDConfig,
		allRefreshETCDConfig,
		restart,
		allETCDNodeHealthCheck,
		refreshETCDConfigToExist,
		allETCDNodeHealthCheck,
	}
	return tasks
}

func handleExistCluster(c *ConfigureModule) []module.Task {

	existETCDHealthCheck := &module.RemoteTask{
		Name:     "ExistETCDHealthCheck",
		Desc:     "Health check on exist etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(NodeETCDExist),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	generateETCDConfig := &module.RemoteTask{
		Name:     "GenerateETCDConfig",
		Desc:     "Generate etcd.env config on new etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(GenerateConfig),
		Parallel: false,
	}

	joinMember := &module.RemoteTask{
		Name:     "JoinETCDMember",
		Desc:     "Join etcd member",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(JoinMember),
		Parallel: false,
	}

	restart := &module.RemoteTask{
		Name:     "RestartETCD",
		Desc:     "Restart etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(RestartETCD),
		Parallel: true,
	}

	newETCDNodeHealthCheck := &module.RemoteTask{
		Name:     "NewETCDNodeHealthCheck",
		Desc:     "Health check on new etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	checkMember := &module.RemoteTask{
		Name:     "CheckETCDMember",
		Desc:     "Check etcd member",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(CheckMember),
		Parallel: true,
	}

	allRefreshETCDConfig := &module.RemoteTask{
		Name:     "AllRefreshETCDConfig",
		Desc:     "Refresh etcd.env config on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(RefreshConfig),
		Parallel: false,
	}

	allETCDNodeHealthCheck := &module.RemoteTask{
		Name:     "AllETCDNodeHealthCheck",
		Desc:     "Health check on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	tasks := []module.Task{
		existETCDHealthCheck,
		generateETCDConfig,
		joinMember,
		restart,
		newETCDNodeHealthCheck,
		checkMember,
		allRefreshETCDConfig,
		allETCDNodeHealthCheck,
	}
	return tasks
}

type BackupModule struct {
	common.KubeModule
}

func (b *BackupModule) Init() {
	b.Name = "ETCDBackupModule"
	b.Desc = "Backup ETCD cluster data"

	backupETCD := &module.RemoteTask{
		Name:     "BackupETCD",
		Desc:     "Backup etcd data regularly",
		Hosts:    b.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(BackupETCD),
		Parallel: true,
	}

	b.Tasks = []module.Task{
		backupETCD,
	}
}
