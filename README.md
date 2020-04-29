# KubeKey
Deploy a Kubernetes Cluster flexibly and easily
## Quick Start
### Check List
Please follow the list to prepare environment.

|  ID   | Check Item  |
|  :----:  | :----  |
|  1  | Require SSH can access to all nodes.  |
|  2  | It's recommended that Your OS is clean (without any other software installed), otherwise there may be conflicts.  |
|  3  | OS requirements (For Minimal Installation of KubeSphere only)：at least 2 vCPUs and 4GB RAM. |
|  4  | Make sure the storage service is available if you want to deploy a cluster with KubeSphere.<br>The relevant client should be installed on all nodes in cluster, if you storage server is [nfs / ceph / glusterfs](./docs/storage-client.md).   |
|  5  | Make sure the DNS address in /etc/resolv.conf is available. Otherwise, it may cause some issues of DNS in cluster. |
|  6  | If your network configuration uses Firewall or Security Group，you must ensure infrastructure components can communicate with each other through specific ports.<br>It's recommended that you turn off the firewall or follow the link configuriation: [NetworkAccess](./docs/network-access.md)|
|  7  | A container image mirror (accelerator) is recommended to be prepared, if you have trouble downloading images from dockerhub.io.  |            

### Usage
* Download binary
```shell script
curl -O -k https://kubernetes.pek3b.qingstor.com/tools/kubekey/kk
chmod +x kk
```
* Deploy a Allinone cluster
```shell script
./kk create cluster
```
* Deploy a MultiNodes cluster
  
> Create a example configuration file by following command or [example configuration file](docs/config-example.md)
```shell script
./kk create config      # Only kubernetes
./kk create config --add localVolume      # Add plugins (eg: localVolume / nfsClient / localVolume,nfsClient)

# Please fill in the configuration file under the current path (k2cluster-example.yaml) according to the environmental information
```
> Deploy cluster
```shell script
./kk create cluster -f ./k2cluster-example.yaml
```
* Add Nodes
> Add new node's information to the cluster config file
```shell script
./kk scale -f ./k2cluster-example.yaml
```
* Reset Cluster
```shell script
# allinone
./kk reset

# multinodes
./kk reset -f ./k2cluster-example.yaml
```
### Supported
* Deploy allinone cluster
* Deploy multinodes cluster
* Add nodes (masters and nodes)

### Build
```shell script
git clone https://github.com/pixiake/kubekey.git
cd kubekey
./build.sh
```
> Note: Docker needs to be installed before building.
## Quick Start
* CaaO (Cluster as a Object)
* Support more container runtimes: cri-o containerd

