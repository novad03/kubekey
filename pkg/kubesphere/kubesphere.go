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

package kubesphere

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"strings"
	"time"
)

var stopChan = make(chan string, 1)

func DeployKubeSphere(mgr *manager.Manager) error {
	if mgr.Cluster.KubeSphere.Console.Port != 0 {
		mgr.Logger.Infoln("Deploying KubeSphere ...")
		if err := mgr.RunTaskOnMasterNodes(deployKubeSphere, true); err != nil {
			return err
		}
		ResultNotes()
	}
	return nil
}

func deployKubeSphere(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	if mgr.Runner.Index == 0 {
		if err := DeployKubeSphereStep(mgr); err != nil {
			return err
		}
	}

	return nil
}

func DeployKubeSphereStep(mgr *manager.Manager) error {
	kubesphereYaml, err := GenerateKubeSphereYaml(mgr)
	if err != nil {
		return err
	}
	kubesphereYamlBase64 := base64.StdEncoding.EncodeToString([]byte(kubesphereYaml))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/addons/kubesphere.yaml\"", kubesphereYamlBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate KubeSphere manifests")
	}
	_, err2 := mgr.Runner.RunCmdOutput("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/kubesphere.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to deploy kubesphere.yaml")
	}

	go CheckKubeSphereStatus(mgr)
	return nil
}

func CheckKubeSphereStatus(mgr *manager.Manager) {
	for i := 30; i > 0; i-- {
		time.Sleep(10 * time.Second)
		_, err := mgr.Runner.RunCmd(
			"/usr/local/bin/kubectl exec -n kubesphere-system $(kubectl get pod -n kubesphere-system -l app=ks-install -o jsonpath='{.items[0].metadata.name}') ls kubesphere/playbooks/kubesphere_running",
		)
		if err == nil {
			output, err := mgr.Runner.RunCmd(
				"/usr/local/bin/kubectl exec -n kubesphere-system $(kubectl get pod -n kubesphere-system -l app=ks-install -o jsonpath='{.items[0].metadata.name}') cat kubesphere/playbooks/kubesphere_running",
			)
			if err == nil && output != "" {
				stopChan <- output
				break
			}
		}
	}
	stopChan <- ""
}

func ResultNotes() {
	var (
		position = 1
		notes    = "Please wait for the installation to complete: "
	)
	fmt.Println("\n")
Loop:
	for {
		select {
		case result := <-stopChan:
			fmt.Printf("\033[%dA\033[K", position)
			fmt.Println(result)
			break Loop
		default:
			for i := 0; i < 10; i++ {
				if i < 5 {
					fmt.Printf("\033[%dA\033[K", position)

					output := fmt.Sprintf(
						"%s%s%s",
						notes,
						strings.Repeat(" ", i),
						">>--->",
					)

					fmt.Printf("%s \033[K\n", output)
					time.Sleep(time.Duration(200) * time.Millisecond)
				} else {
					fmt.Printf("\033[%dA\033[K", position)

					output := fmt.Sprintf(
						"%s%s%s",
						notes,
						strings.Repeat(" ", 10-i),
						"<---<<",
					)

					fmt.Printf("%s \033[K\n", output)
					time.Sleep(time.Duration(200) * time.Millisecond)
				}
			}
		}
	}
}
