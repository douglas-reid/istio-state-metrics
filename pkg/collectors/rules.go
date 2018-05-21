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
	fmt.Println("trying to collect with store: ", rc.store)
	rules, err := rc.store.List()
	if err != nil {
		ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "rule"}).Inc()
		return
	}
	ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "rule"}).Add(0)

	ResourcesPerScrapeMetric.With(prometheus.Labels{"resource": "rule"}).Observe(float64(len(rules)))
	for _, r := range rules {
		rc.collectRule(ch, r)
	}
}

func (rc *ruleCollector) collectRule(ch chan<- prometheus.Metric, r v1alpha2.Rule) {
	addConstMetric := func(desc *prometheus.Desc, t prometheus.ValueType, v float64, lv ...string) {
		lv = append([]string{fmt.Sprintf("%s.%s", r.Name, r.Namespace), r.Spec.Match}, lv...)
		ch <- prometheus.MustNewConstMetric(desc, t, v, lv...)
	}
	addGauge := func(desc *prometheus.Desc, v float64, lv ...string) {
		addConstMetric(desc, prometheus.GaugeValue, v, lv...)
	}

	for _, action := range r.Spec.Actions {
		addGauge(descRuleAction, 1, action.Handler, strings.Join(action.Instances, ","))
	}
}
