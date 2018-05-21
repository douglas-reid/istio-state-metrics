package options

import (
	"github.com/douglas-reid/istio-state-metrics/pkg/client/clientset/versioned"
	"github.com/douglas-reid/istio-state-metrics/pkg/collectors"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DefaultNamespaces = NamespaceList{metav1.NamespaceAll}
	DefaultCollectors = CollectorSet{
		"rules": struct{}{},
	}
	AvailableCollectors = map[string]func(registry prometheus.Registerer, client versioned.Interface, namespaces []string){
		"rules": collectors.RegisterRuleCollector,
	}
)
