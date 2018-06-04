package collectors

import (
	"context"
	"fmt"
	"github.com/douglas-reid/istio-state-metrics/pkg/apis/networking/v1alpha3"
	"github.com/douglas-reid/istio-state-metrics/pkg/client/clientset/versioned"
	informers "github.com/douglas-reid/istio-state-metrics/pkg/client/informers/externalversions/networking/v1alpha3"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
)

var (
	descVirtualServiceHostsName          = "istio_pilot_virtual_service_hosts"
	descVirtualServiceHostsHelp          = "Information about Hosts in Pilot VirtualServices"
	descVirtualServiceHostsDefaultLabels = []string{"VirtualService", "hosts"}

	descVirtualServiceHosts = prometheus.NewDesc(
		descVirtualServiceHostsName,
		descVirtualServiceHostsHelp,
		descVirtualServiceHostsDefaultLabels,
		nil,
	)

	descVirtualServicesInfo = prometheus.NewDesc(
		"istio_pilot_virtualService_info",
		"Information about Pilot VirtualServices",
		[]string{"VirtualService", "namespace"},
		nil,
	)

	descVirtualServicesGateways = prometheus.NewDesc(
		"istio_pilot_virtualService_gateways",
		"Information about Gateways in Pilot VirtualServices",
		[]string{"VirtualService", "gateways"},
		nil,
	)
)

type VirtualServiceLister func() ([]v1alpha3.VirtualService, error)

func (l VirtualServiceLister) List() ([]v1alpha3.VirtualService, error) {
	return l()
}

func RegisterVirtualServiceCollector(registry prometheus.Registerer, client versioned.Interface, namespaces []string) {
	fn := func(ns string) cache.SharedIndexInformer {
		return informers.NewVirtualServiceInformer(client, ns, resyncPeriod, cache.Indexers{})
	}

	sinfs := NewSharedInformerList(fn, namespaces)

	virtualServiceLister := VirtualServiceLister(func() (virtualServices []v1alpha3.VirtualService, err error) {
		for _, sinf := range *sinfs {
			for _, m := range sinf.GetStore().List() {
				virtualServices = append(virtualServices, *m.(*v1alpha3.VirtualService))
			}
		}
		return virtualServices, nil
	})

	registry.MustRegister(&virtualServiceCollector{store: virtualServiceLister})
	sinfs.Run(context.Background().Done())
}

type virtualServiceStore interface {
	List() (virtualServices []v1alpha3.VirtualService, err error)
}

// virtualServicesCollector collects metrics about all virtualServices in the cluster.
type virtualServiceCollector struct {
	store virtualServiceStore
}

// Describe implements the prometheus.Collector interface.
func (rc *virtualServiceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descVirtualServiceHosts
}

// Collect implements the prometheus.Collector interface.
func (rc *virtualServiceCollector) Collect(ch chan<- prometheus.Metric) {
	virtualServices, err := rc.store.List()
	if err != nil {
		ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "VirtualService"}).Inc()
		return
	}
	ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "VirtualService"}).Add(0)

	ResourcesPerScrapeMetric.With(prometheus.Labels{"resource": "VirtualService"}).Observe(float64(len(virtualServices)))
	for _, r := range virtualServices {
		rc.collectVirtualService(ch, r)
	}
}

func (rc *virtualServiceCollector) collectVirtualService(ch chan<- prometheus.Metric, r v1alpha3.VirtualService) {
	for _, host := range r.Spec.Hosts {
		ch <- prometheus.MustNewConstMetric(descVirtualServiceHosts, prometheus.GaugeValue, 1,
			fmt.Sprintf("%s.%s", r.Name, r.Namespace),
			host)
	}

	for _, gateway := range r.Spec.Gateways {
		ch <- prometheus.MustNewConstMetric(descVirtualServicesGateways, prometheus.GaugeValue, 1,
			fmt.Sprintf("%s.%s", r.Name, r.Namespace),
			gateway)
	}

	// virtualService info metric
	ch <- prometheus.MustNewConstMetric(descVirtualServicesInfo, prometheus.GaugeValue, 1, r.Name, r.Namespace)
}
