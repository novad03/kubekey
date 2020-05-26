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

package images

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
)

type Image struct {
	RepoAddr  string
	Namespace string
	Repo      string
	Tag       string
	Group     string
	Enable    bool
}

func (image *Image) ImageName() string {
	var prefix string

	if image.RepoAddr == "" {
		if image.Namespace == "" {
			prefix = ""
		} else {
			prefix = fmt.Sprintf("%s/", image.Namespace)
		}
	} else {
		if image.Namespace == "" {
			prefix = fmt.Sprintf("%s/library/", image.RepoAddr)
		} else {
			prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.Namespace)
		}
	}

	return fmt.Sprintf("%s%s:%s", prefix, image.Repo, image.Tag)
}

func (image *Image) ImageRepo() string {
	var prefix string

	if image.RepoAddr == "" {
		if image.Namespace == "" {
			prefix = ""
		} else {
			prefix = fmt.Sprintf("%s/", image.Namespace)
		}
	} else {
		if image.Namespace == "" {
			prefix = fmt.Sprintf("%s/library/", image.RepoAddr)
		} else {
			prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.Namespace)
		}
	}

	return fmt.Sprintf("%s%s", prefix, image.Repo)
}

func PreDownloadImages(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Start to download images on all nodes")

	return mgr.RunTaskOnAllNodes(preDownloadImages, true)
}

func preDownloadImages(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	imagesList := []*Image{
		GetImage(mgr, "etcd"),
		GetImage(mgr, "pause"),
		GetImage(mgr, "kube-apiserver"),
		GetImage(mgr, "kube-controller-manager"),
		GetImage(mgr, "kube-scheduler"),
		GetImage(mgr, "kube-proxy"),
		GetImage(mgr, "coredns"),
		GetImage(mgr, "k8s-dns-node-cache"),
		GetImage(mgr, "calico-kube-controllers"),
		GetImage(mgr, "calico-cni"),
		GetImage(mgr, "calico-node"),
		GetImage(mgr, "calico-flexvol"),
		GetImage(mgr, "flannel"),
	}

	for _, image := range imagesList {

		if node.IsMaster && image.Group == "master" && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.Name, image.ImageName())
			_, err := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E docker pull %s", image.ImageName()))
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to download image: %s", image.ImageName()))
			}
		}
		if (node.IsMaster || node.IsWorker) && image.Group == "k8s" && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.Name, image.ImageName())
			_, err := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E docker pull %s", image.ImageName()))
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to download image: %s", image.ImageName()))
			}
		}
		if node.IsEtcd && image.Group == "etcd" && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.Name, image.ImageName())
			_, err := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E docker pull %s", image.ImageName()))
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to download image: %s", image.ImageName()))
			}
		}
	}

	return nil
}
