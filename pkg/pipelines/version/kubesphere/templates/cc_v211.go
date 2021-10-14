package templates

import (
	"github.com/lithammer/dedent"
	"text/template"
)

var V211 = template.Must(template.New("v2.1.1").Parse(
	dedent.Dedent(`---
apiVersion: v1
data:
  ks-config.yaml: |
    ---
    local_registry: ""
    persistence:
      storageClass: ""
    etcd:
      monitoring: true
      endpointIps: localhost
      port: 2379
      tlsEnable: true
    common:
      mysqlVolumeSize: 20Gi
      minioVolumeSize: 20Gi
      etcdVolumeSize: 20Gi
      openldapVolumeSize: 2Gi
      redisVolumSize: 2Gi
    metrics_server:
      enabled: false
    console:
      enableMultiLogin: False  # enable/disable multi login
      port: 30880
    monitoring:
      prometheusReplicas: 1
      prometheusMemoryRequest: 400Mi
      prometheusVolumeSize: 20Gi
      grafana:
        enabled: false
    logging:
      enabled: false
      elasticsearchMasterReplicas: 1
      elasticsearchDataReplicas: 1
      logsidecarReplicas: 2
      elasticsearchMasterVolumeSize: 4Gi
      elasticsearchDataVolumeSize: 20Gi
      logMaxAge: 7
      elkPrefix: logstash
      containersLogMountedPath: ""
      kibana:
        enabled: false
    openpitrix:
      enabled: false
    devops:
      enabled: false
      jenkinsMemoryLim: 2Gi
      jenkinsMemoryReq: 1500Mi
      jenkinsVolumeSize: 8Gi
      jenkinsJavaOpts_Xms: 512m
      jenkinsJavaOpts_Xmx: 512m
      jenkinsJavaOpts_MaxRAM: 2g
      sonarqube:
        enabled: false
        postgresqlVolumeSize: 8Gi
    servicemesh:
      enabled: false
    notification:
      enabled: false
    alerting:
      enabled: false
kind: ConfigMap
metadata:
  name: ks-installer
  namespace: kubesphere-system
  labels:
    version: {{ .Tag }}
`)))
