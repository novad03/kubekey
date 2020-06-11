/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package etcd

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/etcd/tmpl"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"strings"
	"time"
)

var (
	certsStr         = make(chan map[string]string)
	certsContent     = map[string]string{}
	etcdBackupPrefix = "/var/backups"
	etcdDataDir      = "/var/lib/etcd"
	etcdConfigDir    = "/etc/ssl/etcd"
	etcdCertDir      = "/etc/ssl/etcd/ssl"
	etcdBinDir       = "/usr/local/bin"
	accessAddresses  = ""
	peerAddresses    = []string{}
	etcdStatus       = ""
)

func GenerateEtcdCerts(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Generating etcd certs")

	return mgr.RunTaskOnEtcdNodes(generateCerts, true)
}

func generateCerts(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {

	if mgr.Runner.Index == 0 {
		certsScript, err := tmpl.GenerateEtcdSslScript(mgr)
		if err != nil {
			return err
		}
		certsScriptBase64 := base64.StdEncoding.EncodeToString([]byte(certsScript))
		_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("echo %s | base64 -d > /tmp/kubekey/make-ssl-etcd.sh && chmod +x /tmp/kubekey/make-ssl-etcd.sh", certsScriptBase64))
		if err1 != nil {
			return errors.Wrap(errors.WithStack(err1), "Failed to generate etcd certs script")
		}
		certsOpensslCfg, err := tmpl.GenerateEtcdSslCfg(mgr.Cluster)
		if err != nil {
			return err
		}
		certsOpensslCfgBase64 := base64.StdEncoding.EncodeToString([]byte(certsOpensslCfg))
		_, err2 := mgr.Runner.RunCmd(fmt.Sprintf("echo %s | base64 -d > /tmp/kubekey/openssl.conf", certsOpensslCfgBase64))
		if err2 != nil {
			return errors.Wrap(errors.WithStack(err2), "Failed to generate etcd certs script")
		}

		cmd := fmt.Sprintf("mkdir -p %s && /bin/bash -x %s/make-ssl-etcd.sh -f %s/openssl.conf -d %s", etcdCertDir, "/tmp/kubekey", "/tmp/kubekey", etcdCertDir)

		_, err3 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", cmd))
		if err3 != nil {
			return errors.Wrap(errors.WithStack(err3), "Failed to generate etcd certs")
		}

		for _, cert := range generateCertsFiles(mgr) {
			certsBase64Cmd := fmt.Sprintf("sudo -E /bin/sh -c \"cat %s/%s | base64 --wrap=0\"", etcdCertDir, cert)
			certsBase64, err4 := mgr.Runner.RunCmd(certsBase64Cmd)
			if err4 != nil {
				return errors.Wrap(errors.WithStack(err4), "Failed to get etcd certs content")
			}
			certsContent[cert] = certsBase64
		}

		for i := 1; i <= len(mgr.EtcdNodes)-1; i++ {
			certsStr <- certsContent
		}

	} else {
		mgr.Runner.RunCmdOutput(fmt.Sprintf("sudo mkdir -p %s", etcdCertDir))
		for file, cert := range <-certsStr {
			writeCertCmd := fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > %s/%s\"", cert, etcdCertDir, file)
			_, err4 := mgr.Runner.RunCmd(writeCertCmd)
			if err4 != nil {
				return errors.Wrap(errors.WithStack(err4), "Failed to write etcd certs content")
			}
		}
	}

	return nil
}

func generateCertsFiles(mgr *manager.Manager) []string {
	var certsList []string
	certsList = append(certsList, "ca.pem")
	certsList = append(certsList, "ca-key.pem")
	for _, host := range mgr.EtcdNodes {
		certsList = append(certsList, fmt.Sprintf("admin-%s.pem", host.Name))
		certsList = append(certsList, fmt.Sprintf("admin-%s-key.pem", host.Name))
		certsList = append(certsList, fmt.Sprintf("member-%s.pem", host.Name))
		certsList = append(certsList, fmt.Sprintf("member-%s-key.pem", host.Name))
	}
	for _, host := range mgr.MasterNodes {
		certsList = append(certsList, fmt.Sprintf("node-%s.pem", host.Name))
		certsList = append(certsList, fmt.Sprintf("node-%s-key.pem", host.Name))
	}
	return certsList
}

func SyncEtcdCertsToMaster(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Synchronizing etcd certs")

	return mgr.RunTaskOnMasterNodes(syncEtcdCertsToMaster, true)
}

func syncEtcdCertsToMaster(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	if !node.IsEtcd {
		mgr.Runner.RunCmdOutput(fmt.Sprintf("sudo mkdir -p %s", etcdCertDir))
		for file, cert := range certsContent {
			writeCertCmd := fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > %s/%s\"", cert, etcdCertDir, file)
			_, err := mgr.Runner.RunCmd(writeCertCmd)
			if err != nil {
				return errors.Wrap(errors.WithStack(err), "Failed to sync etcd certs to master")
			}
		}
	}
	return nil
}

func GenerateEtcdService(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Creating etcd service")

	return mgr.RunTaskOnEtcdNodes(generateEtcdService, true)
}

func generateEtcdService(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	etcdService, err := tmpl.GenerateEtcdService(mgr, mgr.Runner.Index)
	if err != nil {
		return err
	}
	etcdServiceBase64 := base64.StdEncoding.EncodeToString([]byte(etcdService))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/etcd.service\"", etcdServiceBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate etcd service")
	}

	etcdBin, err := tmpl.GenerateEtcdBinary(mgr, mgr.Runner.Index)
	if err != nil {
		return err
	}
	etcdBinBase64 := base64.StdEncoding.EncodeToString([]byte(etcdBin))
	_, err3 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /usr/local/bin/etcd && chmod +x /usr/local/bin/etcd\"", etcdBinBase64))
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "Failed to generate etcd bin")
	}

	getEtcdCtlCmd := fmt.Sprintf("docker run --rm -v /usr/local/bin:/systembindir %s /bin/cp /usr/local/bin/etcdctl /systembindir/etcdctl", preinstall.GetImage(mgr, "etcd").ImageName())
	_, err4 := mgr.Runner.RunCmdOutput(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", getEtcdCtlCmd))
	if err4 != nil {
		return errors.Wrap(errors.WithStack(err4), "Failed to get etcdctl")
	}

	if err := restartEtcd(mgr); err != nil {
		return err
	}

	addrList := []string{}
	for _, host := range mgr.EtcdNodes {
		addrList = append(addrList, fmt.Sprintf("https://%s:2379", host.InternalAddress))
	}

	accessAddresses = strings.Join(addrList, ",")

	return nil
}

func SetupEtcdCluster(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Starting etcd cluster")

	return mgr.RunTaskOnEtcdNodes(setupEtcdCluster, false)
}

func setupEtcdCluster(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	localPeerAddresses := []string{}
	output, _ := mgr.Runner.RunCmdOutput("sudo -E /bin/sh -c \"[ -f /etc/etcd.env ] && echo 'Configuration file already exists' || echo 'Configuration file will be created'\"")
	if strings.TrimSpace(output) == "Configuration file already exists" {
		if err := helthCheck(mgr, node); err != nil {
			return err
		}
		etcdStatus = "existing"
		for i := 0; i <= mgr.Runner.Index; i++ {
			localPeerAddresses = append(localPeerAddresses, fmt.Sprintf("etcd%d=https://%s:2380", i+1, mgr.EtcdNodes[i].InternalAddress))
		}
		if mgr.Runner.Index == len(mgr.EtcdNodes)-1 {
			peerAddresses = localPeerAddresses
		}
	} else {
		for i := 0; i <= mgr.Runner.Index; i++ {
			localPeerAddresses = append(localPeerAddresses, fmt.Sprintf("etcd%d=https://%s:2380", i+1, mgr.EtcdNodes[i].InternalAddress))
		}
		if mgr.Runner.Index == len(mgr.EtcdNodes)-1 {
			peerAddresses = localPeerAddresses
		}
		if mgr.Runner.Index == 0 {
			if err := refreshConfig(mgr, node, mgr.Runner.Index, localPeerAddresses, "new"); err != nil {
				return err
			}
			etcdStatus = "new"
		} else {
			switch etcdStatus {
			case "new":
				if err := refreshConfig(mgr, node, mgr.Runner.Index, localPeerAddresses, "new"); err != nil {
					return err
				}
			case "existing":
				if err := refreshConfig(mgr, node, mgr.Runner.Index, localPeerAddresses, "existing"); err != nil {
					return err
				}
				joinMemberCmd := fmt.Sprintf("sudo -E /bin/sh -c \"export ETCDCTL_API=2;export ETCDCTL_CERT_FILE='/etc/ssl/etcd/ssl/admin-%s.pem';export ETCDCTL_KEY_FILE='/etc/ssl/etcd/ssl/admin-%s-key.pem';export ETCDCTL_CA_FILE='/etc/ssl/etcd/ssl/ca.pem';%s/etcdctl --endpoints=%s member add %s %s\"", node.Name, node.Name, etcdBinDir, accessAddresses, fmt.Sprintf("etcd%d", mgr.Runner.Index+1), fmt.Sprintf("https://%s:2380", node.InternalAddress))
				_, err := mgr.Runner.RunCmdOutput(joinMemberCmd)
				if err != nil {
					return errors.Wrap(errors.WithStack(err), "Failed to add etcd member")
				}
				if err := restartEtcd(mgr); err != nil {
					return err
				}
				if err := helthCheck(mgr, node); err != nil {
					return err
				}
				checkMemberCmd := fmt.Sprintf("sudo -E /bin/sh -c \"export ETCDCTL_API=2;export ETCDCTL_CERT_FILE='/etc/ssl/etcd/ssl/admin-%s.pem';export ETCDCTL_KEY_FILE='/etc/ssl/etcd/ssl/admin-%s-key.pem';export ETCDCTL_CA_FILE='/etc/ssl/etcd/ssl/ca.pem';%s/etcdctl --no-sync --endpoints=%s member list\"", node.Name, node.Name, etcdBinDir, accessAddresses)
				memberList, err := mgr.Runner.RunCmdOutput(checkMemberCmd)
				if err != nil {
					return errors.Wrap(errors.WithStack(err), "Failed to list etcd member")
				}
				if !strings.Contains(memberList, fmt.Sprintf("https://%s:2379", node.InternalAddress)) {
					return errors.Wrap(errors.WithStack(err), "Failed to add etcd member")
				}
			default:
				return errors.New("Failed to get etcd cluster status")
			}
		}

	}
	return nil
}

func RefreshEtcdConfig(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Refreshing etcd configuration")

	return mgr.RunTaskOnEtcdNodes(refreshEtcdConfig, true)
}

func refreshEtcdConfig(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {

	if etcdStatus == "new" {
		if err := refreshConfig(mgr, node, mgr.Runner.Index, peerAddresses, "new"); err != nil {
			return err
		}
		if err := restartEtcd(mgr); err != nil {
			return err
		}
		if err := helthCheck(mgr, node); err != nil {
			return err
		}
	}

	if err := refreshConfig(mgr, node, mgr.Runner.Index, peerAddresses, "existing"); err != nil {
		return err
	}

	if err := helthCheck(mgr, node); err != nil {
		return err
	}

	return nil
}

func helthCheck(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {
	checkHealthCmd := fmt.Sprintf("sudo -E /bin/sh -c \"export ETCDCTL_API=2;export ETCDCTL_CERT_FILE='/etc/ssl/etcd/ssl/admin-%s.pem';export ETCDCTL_KEY_FILE='/etc/ssl/etcd/ssl/admin-%s-key.pem';export ETCDCTL_CA_FILE='/etc/ssl/etcd/ssl/ca.pem';%s/etcdctl --endpoints=%s cluster-health | grep -q 'cluster is healthy'\"", node.Name, node.Name, etcdBinDir, accessAddresses)
helthCheckLoop:
	for i := 20; i > 0; i-- {
		_, err := mgr.Runner.RunCmd(checkHealthCmd)
		if err != nil {
			fmt.Println("Waiting for etcd to start")
			if i == 1 {
				return errors.Wrap(errors.WithStack(err), "Failed to start etcd cluster")
			}
		} else {
			break helthCheckLoop
		}
		time.Sleep(time.Second * 5)
	}
	return nil
}

func refreshConfig(mgr *manager.Manager, node *kubekeyapi.HostCfg, index int, endpoints []string, state string) error {
	etcdEnv, err := tmpl.GenerateEtcdEnv(mgr, node, index, endpoints, state)
	if err != nil {
		return err
	}
	etcdEnvBase64 := base64.StdEncoding.EncodeToString([]byte(etcdEnv))
	_, err2 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/etcd.env\"", etcdEnvBase64))
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to generate etcd env")
	}
	return nil
}

func restartEtcd(mgr *manager.Manager) error {
	_, err5 := mgr.Runner.RunCmdOutput("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl restart etcd && systemctl enable etcd\"")
	if err5 != nil {
		return errors.Wrap(errors.WithStack(err5), "Failed to start etcd")
	}
	return nil
}
