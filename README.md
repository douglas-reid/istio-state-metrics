# istio-state-metrics

## Testing: 
1. Run: `$ go run cmd/istio-state-metrics/main.go --kubeconfig=$HOME/.kube/config`
1. Check: `http://localhost:9090/metrics`

You should see something like:

```
# HELP istio_mixer_instance_info Information about Mixer Instances
# TYPE istio_mixer_instance_info gauge
istio_mixer_instance_info{instance="accesslog",kind="logentry",namespace="istio-system"} 1
istio_mixer_instance_info{instance="appversion",kind="listentry",namespace="istio-system"} 1
istio_mixer_instance_info{instance="attributes",kind="kubernetes",namespace="istio-system"} 1
istio_mixer_instance_info{instance="requestcount",kind="metric",namespace="istio-system"} 1
istio_mixer_instance_info{instance="requestcount",kind="quota",namespace="istio-system"} 1
istio_mixer_instance_info{instance="requestduration",kind="metric",namespace="istio-system"} 1
istio_mixer_instance_info{instance="requestsize",kind="metric",namespace="istio-system"} 1
istio_mixer_instance_info{instance="responsesize",kind="metric",namespace="istio-system"} 1
istio_mixer_instance_info{instance="tcpbytereceived",kind="metric",namespace="istio-system"} 1
istio_mixer_instance_info{instance="tcpbytesent",kind="metric",namespace="istio-system"} 1
# HELP istio_mixer_rule_actions Information about Actions in Mixer Rules
# TYPE istio_mixer_rule_actions gauge
istio_mixer_rule_actions{handler="handler.kubernetesenv",instances="attributes.kubernetes",match="",rule="kubeattrgenrulerule.istio-system"} 1
istio_mixer_rule_actions{handler="handler.kubernetesenv",instances="attributes.kubernetes",match="context.protocol == \"tcp\"",rule="tcpkubeattrgenrulerule.istio-system"} 1
istio_mixer_rule_actions{handler="handler.memquota",instances="requestcount.quota",match="",rule="quota.istio-system"} 1
istio_mixer_rule_actions{handler="handler.prometheus",instances="requestcount.metric,requestduration.metric,requestsize.metric,responsesize.metric",match="",rule="promhttp.istio-system"} 1
istio_mixer_rule_actions{handler="handler.prometheus",instances="tcpbytesent.metric,tcpbytereceived.metric",match="",rule="promtcp.istio-system"} 1
istio_mixer_rule_actions{handler="handler.stdio",instances="accesslog.logentry",match="true",rule="stdio.istio-system"} 1
istio_mixer_rule_actions{handler="staticversion.listchecker",instances="appversion.listentry",match="(destination.labels[\"app\"]|\"unknown\") == \"ratings\"",rule="checkwl.istio-system"} 1
# HELP istio_mixer_rule_info Information about Mixer Rules
# TYPE istio_mixer_rule_info gauge
istio_mixer_rule_info{namespace="istio-system",rule="checkwl"} 1
istio_mixer_rule_info{namespace="istio-system",rule="kubeattrgenrulerule"} 1
istio_mixer_rule_info{namespace="istio-system",rule="promhttp"} 1
istio_mixer_rule_info{namespace="istio-system",rule="promtcp"} 1
istio_mixer_rule_info{namespace="istio-system",rule="quota"} 1
istio_mixer_rule_info{namespace="istio-system",rule="stdio"} 1
istio_mixer_rule_info{namespace="istio-system",rule="tcpkubeattrgenrulerule"} 1
```