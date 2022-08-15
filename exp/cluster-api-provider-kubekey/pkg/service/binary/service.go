/*
 Copyright 2022 The KubeSphere Authors.

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

package binary

import (
	"text/template"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/scope"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file"
)

type Service struct {
	SSHClient     ssh.Interface
	scope         scope.KKInstanceScope
	instanceScope *scope.InstanceScope

	templateFactory func(sshClient ssh.Interface, template *template.Template, data file.Data, dst string) (operation.Template, error)
	kubeadmFactory  func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
	kubeletFactory  func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
	kubecniFactory  func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
	kubectlFactory  func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
}

func NewService(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) *Service {
	return &Service{
		SSHClient:     sshClient,
		scope:         scope,
		instanceScope: instanceScope,
	}
}

func (s *Service) getTemplateService(template *template.Template, data file.Data, dst string) (operation.Template, error) {
	if s.templateFactory != nil {
		return s.templateFactory(s.SSHClient, template, data, dst)
	}
	return file.NewTemplate(s.SSHClient, s.scope.RootFs(), template, data, dst)
}

func (s *Service) getKubeadmService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if s.kubeadmFactory != nil {
		return s.kubeadmFactory(sshClient, version, arch)
	}
	return file.NewKubeadm(sshClient, s.scope.RootFs(), version, arch)
}

func (s *Service) getKubeletService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if s.kubeletFactory != nil {
		return s.kubeletFactory(sshClient, version, arch)
	}
	return file.NewKubelet(sshClient, s.scope.RootFs(), version, arch)
}

func (s *Service) getKubecniService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if s.kubecniFactory != nil {
		return s.kubecniFactory(sshClient, version, arch)
	}
	return file.NewKubecni(sshClient, s.scope.RootFs(), version, arch)
}

func (s *Service) getKubectlService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if s.kubectlFactory != nil {
		return s.kubectlFactory(sshClient, version, arch)
	}
	return file.NewKubectl(sshClient, s.scope.RootFs(), version, arch)
}
