package collectors

import (
	"context"
	"fmt"
	"github.com/douglas-reid/istio-state-metrics/pkg/apis/networking/v1alpha3"
	"github.com/douglas-reid/istio-state-metrics/pkg/client/clientset/versioned"
	informers "github.com/douglas-reid/istio-state-metrics/pkg/client/informers/externalversions/networking/v1alpha3"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
	"strings"
)

var (
	descVirtualServiceHostsName          = "istio_pilot_virtual_service_host"
	descVirtualServiceHostsHelp          = "Information about Hosts in Pilot VirtualServices"
	descVirtualServiceHostsDefaultLabels = []string{"virtual_service", "namespace", "host"}

	descVirtualServiceHosts = prometheus.NewDesc(
		descVirtualServiceHostsName,
		descVirtualServiceHostsHelp,
		descVirtualServiceHostsDefaultLabels,
		nil,
	)

	descVirtualServicesInfo = prometheus.NewDesc(
		"istio_pilot_virtual_service_info",
		"Information about Pilot VirtualServices",
		[]string{"virtual_service", "namespace"},
		nil,
	)

	descVirtualServicesGateways = prometheus.NewDesc(
		"istio_pilot_virtual_service_gateway",
		"Information about Gateways in Pilot VirtualServices",
		[]string{"virtual_service", "namespace", "gateway"},
		nil,
	)

	descVirtualServicesHttpMatchInfo = prometheus.NewDesc(
		"istio_pilot_virtual_service_http_route_info",
		"Information about Pilot VirtualServices Http Route Info",
		[]string{"virtual_service", "namespace", "uri_exact", "uri_prefix", "uri_regex",
			"scheme_exact", "scheme_prefix", "scheme_regex", "method_exact", "method_prefix", "method_regex",
			"authority_exact", "authority_prefix", "authority_regex", "headers", "port", "source_labels", "gateways"},
		nil,
	)

	descVirtualServicesHttpRouteInfo = prometheus.NewDesc(
		"istio_pilot_virtual_service_http_match_info",
		"Information about Pilot VirtualServices Http Match Info",
		[]string{"virtual_service", "namespace", "destination_host",
			"destination_subset", "destination_port_name", "destination_port_number", "weight"},
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
		ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "virtual_service"}).Inc()
		return
	}
	ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "virtual_service"}).Add(0)

	ResourcesPerScrapeMetric.With(prometheus.Labels{"resource": "virtual_service"}).Observe(float64(len(virtualServices)))
	for _, r := range virtualServices {
		rc.collectVirtualService(ch, r)
	}
}

func (rc *virtualServiceCollector) collectVirtualService(ch chan<- prometheus.Metric, r v1alpha3.VirtualService) {
	for _, host := range r.Spec.Hosts {
		ch <- prometheus.MustNewConstMetric(descVirtualServiceHosts, prometheus.GaugeValue, 1,
			r.Name,
			r.Namespace,
			host)
	}

	for _, gateway := range r.Spec.Gateways {
		ch <- prometheus.MustNewConstMetric(descVirtualServicesGateways, prometheus.GaugeValue, 1,
			r.Name,
			r.Namespace,
			gateway)
	}

	if r.Spec.Http != nil {
		for _, v := range r.Spec.Http {
			if v.Route != nil {
				for _, w := range v.Route {
					destinationHost,
						destinationSubset, destinationPortName, destinationPortNumber, weight := getDestinationWeightVars(w)
					ch <- prometheus.MustNewConstMetric(descVirtualServicesHttpRouteInfo, prometheus.GaugeValue, 1,
						r.Name,
						r.Namespace,
						destinationHost,
						destinationSubset,
						destinationPortName,
						destinationPortNumber,
						weight)
				}
			}

			if v.Match != nil {
				for _, w := range v.Match {
					uriExact, uriPrefix, uriRegex,
						schemeExact, schemePrefix, schemeRegex, methodExact, methodPrefix, methodRegex,
						authorityExact, authorityPrefix, authorityRegex, headers, port, sourceLabels, gateways := getHttpMatchVars(w)
					ch <- prometheus.MustNewConstMetric(descVirtualServicesHttpMatchInfo, prometheus.GaugeValue, 1,
						r.Name,
						r.Namespace,
						uriExact,
						uriPrefix,
						uriRegex,
						schemeExact,
						schemePrefix,
						schemeRegex,
						methodExact,
						methodPrefix,
						methodRegex,
						authorityExact,
						authorityPrefix,
						authorityRegex,
						headers,
						port,
						sourceLabels,
						gateways)
				}
			}
		}
	}

	// virtualService info metric
	ch <- prometheus.MustNewConstMetric(descVirtualServicesInfo, prometheus.GaugeValue, 1, r.Name, r.Namespace)
}

func getDestinationWeightVars(route *v1alpha3.DestinationWeight) (destinationHost string, destinationSubset string, destinationPortName string,
	destinationPortNumber string, weight string) {
	if route == nil {
		return
	}

	destinationHost, destinationSubset, destinationPortName, destinationPortNumber = getDestinationVars(route.Destination)
	weight = fmt.Sprintf("%v", route.Weight)
	return

}

func getDestinationVars(destination *v1alpha3.Destination) (destinationHost string, destinationSubset string, destinationPortName string,
	destinationPortNumber string) {
	if destination == nil {
		return
	}

	destinationHost = destination.Host
	destinationSubset = destination.Subset
	destinationPortName, destinationPortNumber = getPortSelectorVars(destination.Port)
	return
}

func getPortSelectorVars(selector *v1alpha3.PortSelector) (destinationPortName string, destinationPortNumber string) {
	if selector == nil {
		return
	}

	destinationPortName = selector.Name
	destinationPortNumber = fmt.Sprintf("%v", selector.Number)
	return
}

func getHttpMatchVars(request *v1alpha3.HTTPMatchRequest) (uriExact string, uriPrefix string, uriRegex string, schemeExact string,
	schemePrefix string, schemeRegex string, methodExact string, methodPrefix string, methodRegex string, authorityExact string,
	authorityPrefix string, authorityRegex string, headers string, port string, sourceLabels string, gateways string) {
	if request == nil {
		return
	}

	uriExact, uriPrefix, uriRegex = getStringMatchVars(request.Uri)
	schemeExact, schemePrefix, schemeRegex = getStringMatchVars(request.Scheme)
	methodExact, methodPrefix, methodRegex = getStringMatchVars(request.Method)
	authorityExact, authorityPrefix, authorityRegex = getStringMatchVars(request.Authority)

	for k, v := range request.Headers {
		exact, prefix, regex := getStringMatchVars(&v)
		headers += fmt.Sprintf("{header:%s,value:%s,%s,%s}", k, exact, prefix, regex)
	}

	port = fmt.Sprintf("%v", request.Port)

	for k, v := range request.SourceLabels {
		sourceLabels += fmt.Sprintf("{key:%s,value:%s}", k, v)
	}

	gateways = strings.Join(request.Gateways, ",")

	return
}

func getStringMatchVars(stringMatch *v1alpha3.StringMatch) (exact string, prefix string, regex string) {
	if stringMatch == nil {
		return
	}

	exact = stringMatch.Exact
	prefix = stringMatch.Prefix
	regex = stringMatch.Regex
	return
}
