# istio-state-metrics

## Testing: 
1. Run: `$ go run cmd/istio-state-metrics/main.go --kubeconfig=$HOME/.kube/config`
1. Check: `http://localhost:9090/metrics`

You should see something like:

```
# HELP istio_mixer_rule_actions Information about Actions in Mixer Rules
# TYPE istio_mixer_rule_actions gauge
istio_mixer_rule_actions{handler="handler.kubernetesenv",instances="attributes.kubernetes",match="",rule="kubeattrgenrulerule.istio-system"} 1
istio_mixer_rule_actions{handler="handler.kubernetesenv",instances="attributes.kubernetes",match="context.protocol == \"tcp\"",rule="tcpkubeattrgenrulerule.istio-system"} 1
istio_mixer_rule_actions{handler="handler.memquota",instances="requestcount.quota",match="",rule="quota.istio-system"} 1
istio_mixer_rule_actions{handler="handler.prometheus",instances="requestcount.metric,requestduration.metric,requestsize.metric,responsesize.metric",match="",rule="promhttp.istio-system"} 1
istio_mixer_rule_actions{handler="handler.prometheus",instances="tcpbytesent.metric,tcpbytereceived.metric",match="",rule="promtcp.istio-system"} 1
istio_mixer_rule_actions{handler="handler.stdio",instances="accesslog.logentry",match="true",rule="stdio.istio-system"} 1
```