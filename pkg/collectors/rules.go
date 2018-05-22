package collectors

import (
	"context"
	"fmt"
	"strings"

	"github.com/douglas-reid/istio-state-metrics/pkg/apis/config/v1alpha2"
	"github.com/douglas-reid/istio-state-metrics/pkg/client/clientset/versioned"
	informers "github.com/douglas-reid/istio-state-metrics/pkg/client/informers/externalversions/config/v1alpha2"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
)

var (
	descRuleActionName          = "istio_mixer_rule_actions"
	descRuleActionHelp          = "Information about Actions in Mixer Rules"
	descRuleActionDefaultLabels = []string{"rule", "match", "handler", "instances"}

	descRuleAction = prometheus.NewDesc(
		descRuleActionName,
		descRuleActionHelp,
		descRuleActionDefaultLabels,
		nil,
	)

	descRulesInfo = prometheus.NewDesc(
		"istio_mixer_rule_info",
		"Information about Mixer Rules",
		[]string{"rule", "namespace"},
		nil,
	)

	descInstanceInfo = prometheus.NewDesc(
		"istio_mixer_instance_info",
		"Information about Mixer Instances",
		[]string{"instance", "kind", "namespace"},
		nil,
	)
)

type RuleLister func() ([]v1alpha2.Rule, error)

func (l RuleLister) List() ([]v1alpha2.Rule, error) {
	return l()
}

func RegisterRuleCollector(registry prometheus.Registerer, client versioned.Interface, namespaces []string) {
	fn := func(ns string) cache.SharedIndexInformer {
		return informers.NewRuleInformer(client, ns, resyncPeriod, cache.Indexers{})
	}

	sinfs := NewSharedInformerList(fn, namespaces)

	ruleLister := RuleLister(func() (rules []v1alpha2.Rule, err error) {
		for _, sinf := range *sinfs {
			for _, m := range sinf.GetStore().List() {
				rules = append(rules, *m.(*v1alpha2.Rule))
			}
		}
		return rules, nil
	})

	registry.MustRegister(&ruleCollector{store: ruleLister})
	sinfs.Run(context.Background().Done())
}

type ruleStore interface {
	List() (rules []v1alpha2.Rule, err error)
}

// rulesCollector collects metrics about all rules in the cluster.
type ruleCollector struct {
	store ruleStore
}

// Describe implements the prometheus.Collector interface.
func (rc *ruleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descRuleAction
}

// Collect implements the prometheus.Collector interface.
func (rc *ruleCollector) Collect(ch chan<- prometheus.Metric) {
	rules, err := rc.store.List()
	if err != nil {
		ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "rule"}).Inc()
		return
	}
	ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "rule"}).Add(0)

	ResourcesPerScrapeMetric.With(prometheus.Labels{"resource": "rule"}).Observe(float64(len(rules)))
	instances := make(map[string]instanceInfo, len(rules))
	for _, r := range rules {
		rc.collectRule(ch, r)
		newMap := instancesFromRule(r)
		for k, v := range newMap {
			instances[k] = v
		}
	}
	rc.collectInstances(ch, instances)
}

func (rc *ruleCollector) collectRule(ch chan<- prometheus.Metric, r v1alpha2.Rule) {
	for _, action := range r.Spec.Actions {
		ch <- prometheus.MustNewConstMetric(descRuleAction, prometheus.GaugeValue, 1,
			fmt.Sprintf("%s.%s", r.Name, r.Namespace),
			r.Spec.Match,
			action.Handler,
			strings.Join(action.Instances, ","))
	}

	// rule info metric
	ch <- prometheus.MustNewConstMetric(descRulesInfo, prometheus.GaugeValue, 1, r.Name, r.Namespace)
}

func (rc *ruleCollector) collectInstances(ch chan<- prometheus.Metric, instances map[string]instanceInfo) {
	for _, v := range instances {
		ch <- prometheus.MustNewConstMetric(descInstanceInfo, prometheus.GaugeValue, 1, v.name, v.kind, v.namespace)
	}
}

type instanceInfo struct {
	name, kind, namespace string
}

func instancesFromRule(r v1alpha2.Rule) map[string]instanceInfo {
	instances := make(map[string]instanceInfo, len(r.Spec.Actions))
	for _, action := range r.Spec.Actions {
		for _, inst := range action.Instances {
			parts := strings.SplitN(inst, ".", 3)

			ii := instanceInfo{name: inst, kind: "unknown", namespace: r.Namespace}
			for i, part := range parts {
				if i == 0 {
					ii.name = part
					continue
				}
				if i == 1 {
					ii.kind = part
					continue
				}
				ii.namespace = part
			}

			instances[inst] = ii
		}
	}
	return instances
}
