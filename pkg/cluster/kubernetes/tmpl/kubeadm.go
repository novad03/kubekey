package tmpl

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"text/template"
)

var KubeadmCfgTempl = template.Must(template.New("kubeadmCfg").Parse(
	dedent.Dedent(`---
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
etcd:
  external:
    endpoints:
    {{- range .ExternalEtcd.Endpoints }}
    - {{ . }}
    {{- end }}
    caFile: {{ .ExternalEtcd.CaFile }}
    certFile: {{ .ExternalEtcd.CertFile }}
    keyFile: {{ .ExternalEtcd.KeyFile }}
dns:
  type: CoreDNS
  imageRepository: coredns
  imageTag: 1.6.0
imageRepository: {{ .ImageRepo }}
kubernetesVersion: {{ .Version }}
certificatesDir: /etc/kubernetes/pki
clusterName: {{ .ClusterName }}
controlPlaneEndpoint: {{ .ControlPlaneEndpoint }}
networking:
  dnsDomain: {{ .ClusterName }}
  podSubnet: {{ .PodSubnet }}
  serviceSubnet: {{ .ServiceSubnet }}
apiServer:
  extraArgs:
    authorization-mode: Node,RBAC
  timeoutForControlPlane: 4m0s
  certSANs:
    {{- range .CertSANs }}
    - {{ . }}
    {{- end }}

---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
bindAddress: 0.0.0.0
clientConnection:
 acceptContentTypes: 
 burst: 10
 contentType: application/vnd.kubernetes.protobuf
 kubeconfig: 
 qps: 5
clusterCIDR: {{ .PodSubnet }}
configSyncPeriod: 15m0s
conntrack:
 maxPerCore: 32768
 min: 131072
 tcpCloseWaitTimeout: 1h0m0s
 tcpEstablishedTimeout: 24h0m0s
enableProfiling: False
healthzBindAddress: 0.0.0.0:10256
iptables:
 masqueradeAll: False
 masqueradeBit: 14
 minSyncPeriod: 0s
 syncPeriod: 30s
ipvs:
 excludeCIDRs: []
 minSyncPeriod: 0s
 scheduler: rr
 syncPeriod: 30s
 strictARP: False
metricsBindAddress: 127.0.0.1:10249
mode: ipvs
nodePortAddresses: []
oomScoreAdj: -999
portRange: 
udpIdleTimeout: 250ms

---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
clusterDomain: {{ .ClusterName }}
clusterDNS:
- {{ .ClusterIP }}
rotateCertificates: true
kubeReserved:
  cpu: 200m
  memory: 250Mi
systemReserved:
  cpu: 200m
  memory: 250Mi
evictionHard:
  memory.available: 5%
evictionSoft:
  memory.available: 10%
evictionSoftGracePeriod: 
  memory.available: 2m
evictionMaxPodGracePeriod: 120
evictionPressureTransitionPeriod: 30s
    `)))

func GenerateKubeadmCfg(mgr *manager.Manager) (string, error) {
	var externalEtcd kubekeyapi.ExternalEtcd
	var endpointsList []string
	var caFile, certFile, keyFile string

	for _, host := range mgr.EtcdNodes {
		endpoint := fmt.Sprintf("https://%s:%s", host.InternalAddress, kubekeyapi.DefaultEtcdPort)
		endpointsList = append(endpointsList, endpoint)
	}
	externalEtcd.Endpoints = endpointsList

	caFile = "/etc/ssl/etcd/ssl/ca.pem"
	certFile = fmt.Sprintf("/etc/ssl/etcd/ssl/admin-%s.pem", mgr.EtcdNodes[0].Name)
	keyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/admin-%s-key.pem", mgr.EtcdNodes[0].Name)

	externalEtcd.CaFile = caFile
	externalEtcd.CertFile = certFile
	externalEtcd.KeyFile = keyFile

	var imageRepo string
	if mgr.Cluster.Registry.PrivateRegistry != "" {
		imageRepo = fmt.Sprintf("%s/%s", mgr.Cluster.Registry.PrivateRegistry, mgr.Cluster.Kubernetes.ImageRepo)
	} else {
		imageRepo = mgr.Cluster.Kubernetes.ImageRepo
	}
	return util.Render(KubeadmCfgTempl, util.Data{
		"ImageRepo":            imageRepo,
		"Version":              mgr.Cluster.Kubernetes.Version,
		"ClusterName":          mgr.Cluster.Kubernetes.ClusterName,
		"ControlPlaneEndpoint": fmt.Sprintf("%s:%s", mgr.Cluster.ControlPlaneEndpoint.Domain, mgr.Cluster.ControlPlaneEndpoint.Port),
		"PodSubnet":            mgr.Cluster.Network.KubePodsCIDR,
		"ServiceSubnet":        mgr.Cluster.Network.KubeServiceCIDR,
		"CertSANs":             mgr.Cluster.GenerateCertSANs(),
		"ExternalEtcd":         externalEtcd,
		"ClusterIP":            "169.254.25.10",
	})
}
