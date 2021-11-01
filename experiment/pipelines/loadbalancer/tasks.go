package loadbalancer

import (
	"fmt"
	"github.com/kubesphere/kubekey/experiment/core/action"
	"github.com/kubesphere/kubekey/experiment/core/logger"
	"github.com/kubesphere/kubekey/experiment/core/prepare"
	"github.com/kubesphere/kubekey/experiment/core/vars"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"strings"
)

type haproxyPreparatoryWork struct {
	action.BaseAction
}

func (h *haproxyPreparatoryWork) Execute(vars vars.Vars) error {
	if err := h.Runtime.Runner.MkDir("/etc/kubekey/haproxy"); err != nil {
		return err
	}
	if err := h.Runtime.Runner.Chmod("/etc/kubekey/haproxy", os.FileMode(0777)); err != nil {
		return err
	}
	return nil
}

type getChecksum struct {
	action.BaseAction
}

func (g *getChecksum) Execute(vars vars.Vars) error {
	md5Str, err := g.Runtime.Runner.FileMd5("/etc/kubekey/haproxy/haproxy.cfg")
	if err != nil {
		return err
	}
	g.Cache.Set("md5", md5Str)
	return nil
}

type updateK3sPrepare struct {
	prepare.BasePrepare
}

func (u *updateK3sPrepare) PreCheck() (bool, error) {
	exist, err := u.Runtime.Runner.FileExist("/etc/systemd/system/k3s.service")
	if err != nil {
		return false, err
	}

	if exist {
		if out, err := u.Runtime.Runner.SudoCmd("sed -n '/--server=.*/p' /etc/systemd/system/k3s.service", false); err != nil {
			return false, err
		} else {
			if strings.Contains(strings.TrimSpace(out), LocalServer) {
				logger.Log.Debugf("do not restart kubelet, /etc/systemd/system/k3s.service content is %s", out)
				return false, nil
			}
		}
	} else {
		return false, errors.New("Failed to find /etc/systemd/system/k3s.service")
	}
	return true, nil
}

type updateK3s struct {
	action.BaseAction
}

func (u *updateK3s) Execute(vars vars.Vars) error {
	if _, err := u.Runtime.Runner.SudoCmd("sed -i 's#--server=.*\"#--server=https://127.0.0.1:%s\"#g' /etc/systemd/system/k3s.service", false); err != nil {
		return err
	}
	if _, err := u.Runtime.Runner.SudoCmd("systemctl restart k3s", false); err != nil {
		return err
	}
	return nil
}

type updateKubeletPrepare struct {
	prepare.BasePrepare
}

func (u *updateKubeletPrepare) PreCheck() (bool, error) {
	exist, err := u.Runtime.Runner.FileExist("/etc/kubernetes/kubelet.conf")
	if err != nil {
		return false, err
	}

	if exist {
		if out, err := u.Runtime.Runner.SudoCmd("sed -n '/server:.*/p' /etc/kubernetes/kubelet.conf", true); err != nil {
			return false, err
		} else {
			if strings.Contains(strings.TrimSpace(out), LocalServer) {
				logger.Log.Debugf("do not restart kubelet, /etc/kubernetes/kubelet.conf content is %s", out)
				return false, nil
			}
		}
	} else {
		return false, errors.New("Failed to find /etc/kubernetes/kubelet.conf")
	}
	return true, nil
}

type updateKubelet struct {
	action.BaseAction
}

func (u *updateKubelet) Execute(vars vars.Vars) error {
	if _, err := u.Runtime.Runner.SudoCmd(fmt.Sprintf(
		"sed -i 's#server:.*#server: https://127.0.0.1:%s#g' /etc/kubernetes/kubelet.conf",
		strconv.Itoa(u.Runtime.Cluster.ControlPlaneEndpoint.Port)), false); err != nil {
		return err
	}
	if _, err := u.Runtime.Runner.SudoCmd("systemctl daemon-reload && systemctl restart kubelet", false); err != nil {
		return err
	}
	return nil
}

type updateKubeproxyPrapre struct {
	prepare.BasePrepare
}

func (u *updateKubeproxyPrapre) PreCheck() (bool, error) {
	if out, err := u.Runtime.Runner.SudoCmd(
		"set -o pipefail && /usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf get configmap kube-proxy -n kube-system -o yaml "+
			"| sed -n '/server:.*/p'", false); err != nil {
		return false, err
	} else {
		if strings.Contains(strings.TrimSpace(out), LocalServer) {
			logger.Log.Debugf("do not restart kube-proxy, configmap kube-proxy content is %s", out)
			return false, nil
		}
	}
	return true, nil
}

type updateKubeproxy struct {
	action.BaseAction
}

func (u *updateKubeproxy) Execute(vars vars.Vars) error {
	if _, err := u.Runtime.Runner.SudoCmd("set -o pipefail "+
		"&& /usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf get configmap kube-proxy -n kube-system -o yaml "+
		"| sed 's#server:.*#server: https://127.0.0.1:%s#g' "+
		"| /usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf replace -f -", false); err != nil {
		return err
	}
	if _, err := u.Runtime.Runner.SudoCmd("/usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf delete pod -n kube-system -l k8s-app=kube-proxy --force --grace-period=0", false); err != nil {
		return err
	}
	return nil
}

type updateHosts struct {
	action.BaseAction
}

func (u *updateHosts) Execute(vars vars.Vars) error {
	if _, err := u.Runtime.Runner.SudoCmd("sed -i 's#.* %s#127.0.0.1 %s#g' /etc/hosts", false); err != nil {
		return err
	}
	return nil
}
