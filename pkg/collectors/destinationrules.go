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
	descDestinationRulesInfo = prometheus.NewDesc(
		"istio_pilot_destination_rule_info",
		"Information about Pilot DestinationRules",
		[]string{"destination_rule", "namespace"},
		nil,
	)

	descDestinationRulesHost = prometheus.NewDesc(
		"istio_pilot_destination_rule_host",
		"Information about Host in Pilot DestinationRules",
		[]string{"destination_rule", "host"},
		nil,
	)

	descDestinationRulesTrafficPolicyLoadBalancer = prometheus.NewDesc(
		"istio_pilot_destination_rule_traffic_policy_loadbalancer",
		"Information about LoadBalancer in Pilot DestinationRules",
		[]string{"destination_rule", "lb_type", "lb_identifier", "consistent_hash_minimum_ring_size"},
		nil,
	)

	descDestinationRulesTrafficPolicyTcpConnectionPoolSettings = prometheus.NewDesc(
		"istio_pilot_destination_rule_traffic_policy_tcp_connection_pool_settings",
		"Information about ConnectionPoolSettings in Pilot DestinationRules",
		[]string{"destination_rule", "max_connections", "connect_timeout"},
		nil,
	)
	
	descDestinationRulesTrafficPolicyHttpConnectionPoolSettings = prometheus.NewDesc(
		"istio_pilot_destination_rule_traffic_policy_http_connection_pool_settings",
		"Information about ConnectionPoolSettings in Pilot DestinationRules",
		[]string{"destination_rule", "http1_max_pending_requests",
			"http2_max_requests", "max_requests_per_connection", "max_retries"},
		nil,
	)

	descDestinationRulesTrafficPolicyOutlierDetection = prometheus.NewDesc(
		"istio_pilot_destination_rule_traffic_policy_outlier_detection",
		"Information about OutlierDetection in Pilot DestinationRules",
		[]string{"destination_rule", "consecutive_errors",
			"interval", "base_ejection_time", "max_ejection_percent"},
		nil,
	)

	descDestinationRulesTrafficPolicyTlsSetting = prometheus.NewDesc(
		"istio_pilot_destination_rule_traffic_policy_tls_settings",
		"Information about TLS Settings of TrafficPolicy in Pilot DestinationRules",
		[]string{"destination_rule", "mode", "client_certificate", "private_key", "ca_certificates", "subject_alt_names", "sni"},
		nil,
	)

	descDestinationRulesTrafficPolicyPortTrafficPolicy = prometheus.NewDesc(
		"istio_pilot_destination_rule_traffic_policy_port_level_settings",
		"Information about PortTrafficPolicy in Pilot DestinationRules",
		[]string{"destination_rule", "port_name", "port_number",
			"lb_type", "lb_identifier", "lb_consistent_hash_minimum_ring_size", "tcp_max_connections", "tcp_connect_timeout", "http_http1_max_pending_requests",
			"http_http2_max_requests", "http_max_requests_per_connection", "http_max_retries", "http_consecutive_errors",
			"http_interval", "http_base_ejection_time", "http_max_ejection_percent", "mode", "client_certificate", "private_key", "ca_certificates", "subject_alt_names", "sni"},
		nil,
	)
)

type DestinationRuleLister func() ([]v1alpha3.DestinationRule, error)

func (l DestinationRuleLister) List() ([]v1alpha3.DestinationRule, error) {
	return l()
}

func RegisterDestinationRuleCollector(registry prometheus.Registerer, client versioned.Interface, namespaces []string) {
	fn := func(ns string) cache.SharedIndexInformer {
		return informers.NewDestinationRuleInformer(client, ns, resyncPeriod, cache.Indexers{})
	}

	sinfs := NewSharedInformerList(fn, namespaces)

	destinationRuleLister := DestinationRuleLister(func() (destinationRules []v1alpha3.DestinationRule, err error) {
		for _, sinf := range *sinfs {
			for _, m := range sinf.GetStore().List() {
				destinationRules = append(destinationRules, *m.(*v1alpha3.DestinationRule))
			}
		}
		return destinationRules, nil
	})

	registry.MustRegister(&destinationRuleCollector{store: destinationRuleLister})
	sinfs.Run(context.Background().Done())
}

type destinationRuleStore interface {
	List() (destinationRules []v1alpha3.DestinationRule, err error)
}

// destinationRulesCollector collects metrics about all destinationRules in the cluster.
type destinationRuleCollector struct {
	store destinationRuleStore
}

// Describe implements the prometheus.Collector interface.
func (rc *destinationRuleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descDestinationRulesInfo
}

// Collect implements the prometheus.Collector interface.
func (rc *destinationRuleCollector) Collect(ch chan<- prometheus.Metric) {
	destinationRules, err := rc.store.List()
	if err != nil {
		ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "destination_rule"}).Inc()
		return
	}
	ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "destination_rule"}).Add(0)

	ResourcesPerScrapeMetric.With(prometheus.Labels{"resource": "destination_rule"}).Observe(float64(len(destinationRules)))
	for _, r := range destinationRules {
		rc.collectDestinationRule(ch, r)
	}
}

func (rc *destinationRuleCollector) collectDestinationRule(ch chan<- prometheus.Metric, r v1alpha3.DestinationRule) {
	// destinationRule info metric
	ch <- prometheus.MustNewConstMetric(descDestinationRulesInfo, prometheus.GaugeValue, 1, r.Name, r.Namespace)
	ch <- prometheus.MustNewConstMetric(descDestinationRulesHost, prometheus.GaugeValue, 1,
		fmt.Sprintf("%s.%s", r.Name, r.Namespace),
		r.Spec.Host)
	if r.Spec.TrafficPolicy != nil {
		if r.Spec.TrafficPolicy.Tls != nil {
			mode, clientCertificate, privateKey, caCertificates, subjectAltNames, sni := getTlsSettingsVars(r.Spec.TrafficPolicy.Tls)
			ch <- prometheus.MustNewConstMetric(descDestinationRulesTrafficPolicyTlsSetting, prometheus.GaugeValue, 1,
				fmt.Sprintf("%s.%s", r.Name, r.Namespace),
				mode, clientCertificate, privateKey, caCertificates, subjectAltNames, sni,
			)
		}

		if r.Spec.TrafficPolicy.ConnectionPool != nil {
			tcpMaxConnections, tcpConnectTimeout := getTcpConnectionPoolVars(r.Spec.TrafficPolicy.ConnectionPool)
			ch <- prometheus.MustNewConstMetric(descDestinationRulesTrafficPolicyTcpConnectionPoolSettings, prometheus.GaugeValue, 1,
				fmt.Sprintf("%s.%s", r.Name, r.Namespace),
				tcpMaxConnections, tcpConnectTimeout,
			)
			
			httpHttp1MaxPendingRequests, httpHttp2MaxRequests,
				httpMaxRequestsPerConnection, httpMaxRetries := getConnectionPoolVars(r.Spec.TrafficPolicy.ConnectionPool)
			ch <- prometheus.MustNewConstMetric(descDestinationRulesTrafficPolicyHttpConnectionPoolSettings, prometheus.GaugeValue, 1,
				fmt.Sprintf("%s.%s", r.Name, r.Namespace),
				tcpMaxConnections, httpHttp2MaxRequests,
				httpMaxRequestsPerConnection, httpMaxRetries,
			)
		}

		if r.Spec.TrafficPolicy.LoadBalancer != nil {
			lbType, lbIdentifier, consistentHashMinimumRingSize := getLoadBalancerVars(r.Spec.TrafficPolicy.LoadBalancer)
			ch <- prometheus.MustNewConstMetric(descDestinationRulesTrafficPolicyLoadBalancer, prometheus.GaugeValue, 1,
				fmt.Sprintf("%s.%s", r.Name, r.Namespace),
				lbType, lbIdentifier, consistentHashMinimumRingSize,
			)
		}

		if r.Spec.TrafficPolicy.OutlierDetection != nil {
			httpConsecutiveErrors,
				httpInterval, httpBaseEjectionTime, httpMaxEjectionPercent := getOutlierDetectionVars(r.Spec.TrafficPolicy.OutlierDetection)
			ch <- prometheus.MustNewConstMetric(descDestinationRulesTrafficPolicyOutlierDetection, prometheus.GaugeValue, 1,
				fmt.Sprintf("%s.%s", r.Name, r.Namespace),
				httpConsecutiveErrors, httpInterval, httpBaseEjectionTime, httpMaxEjectionPercent,
			)
		}

		if r.Spec.TrafficPolicy.PortLevelSettings != nil {
			for _, v := range r.Spec.TrafficPolicy.PortLevelSettings {
				portName, portNumber := getPortInfo(v.Port)
				lbType, lbIdentifier, consistentHashMinimumRingSize := getLoadBalancerVars(v.LoadBalancer)
				tcpMaxConnections, tcpConnectTimeout, httpHttp1MaxPendingRequests, httpHttp2MaxRequests,
					httpMaxRequestsPerConnection, httpMaxRetries := getConnectionPoolVars(v.ConnectionPool)
				httpConsecutiveErrors,
					httpInterval, httpBaseEjectionTime, httpMaxEjectionPercent := getOutlierDetectionVars(v.OutlierDetection)
				mode, clientCertificate, privateKey, caCertificates, subjectAltNames, sni := getTlsSettingsVars(v.Tls)
				ch <- prometheus.MustNewConstMetric(descDestinationRulesTrafficPolicyPortTrafficPolicy, prometheus.GaugeValue, 1,
					fmt.Sprintf("%s.%s", r.Name, r.Namespace),
					portName, portNumber,
					lbType, lbIdentifier, consistentHashMinimumRingSize,
					tcpMaxConnections, tcpConnectTimeout, httpHttp1MaxPendingRequests, httpHttp2MaxRequests,
					httpMaxRequestsPerConnection, httpMaxRetries,
					httpConsecutiveErrors, httpInterval, httpBaseEjectionTime, httpMaxEjectionPercent,
					mode, clientCertificate, privateKey, caCertificates, subjectAltNames, sni,
				)
			}
		}
	}
}

func getTlsSettingsVars(tl *v1alpha3.TLSSettings) (mode string, clientCertificate string, privateKey string,
	caCertificates string, subjectAltNames string, sni string) {
	if tl == nil {
		return
	}

	mode = tl.Mode
	clientCertificate = tl.ClientCertificate
	privateKey = tl.PrivateKey
	caCertificates = tl.CaCertificates
	subjectAltNames = strings.Join(tl.SubjectAltNames, ",")
	sni = tl.Sni
	return
}

func getTcpConnectionPoolVars(cp *v1alpha3.ConnectionPoolSettings) (tcpMaxConnections string, tcpConnectTimeout string) {
	if cp == nil {
		return
	}

	if cp.Tcp != nil {
		tcpMaxConnections = fmt.Sprintf("%v", cp.Tcp.MaxConnections)
		tcpConnectTimeout = cp.Tcp.ConnectTimeout.String()
	}
	return
}

func getHttpConnectionPoolVars(cp *v1alpha3.ConnectionPoolSettings) ( httpHttp1MaxPendingRequests string, httpHttp2MaxRequests string, httpMaxRequestsPerConnection string, httpMaxRetries string) {
	if cp == nil {
		return
	}

	if cp.Http != nil {
		httpHttp1MaxPendingRequests = fmt.Sprintf("%v", cp.Http.Http1MaxPendingRequests)
		httpHttp2MaxRequests = fmt.Sprintf("%v", cp.Http.Http2MaxRequests)
		httpMaxRequestsPerConnection = fmt.Sprintf("%v", cp.Http.MaxRequestsPerConnection)
		httpMaxRetries = fmt.Sprintf("%v", cp.Http.MaxRetries)

	}
	return
}

func getLoadBalancerVars(lb *v1alpha3.LoadBalancerSettings) (lbType string, lbIdentifier string, consistentHashMinimumRingSize string) {
	if lb == nil {
		return
	}

	/*if x, ok := lb.GetLbPolicy().(*v1alpha3.LoadBalancerSettings_Simple); ok {
		lbType ="simple"
		lbIdentifier = x.Simple.String()
	}*/
	if lb.ConsistentHash != nil {
		//if x, ok := lb.GetLbPolicy().(*v1alpha3.LoadBalancerSettings_ConsistentHash); ok {
		lbType = "consistent_hash"
		lbIdentifier = lb.ConsistentHash.HttpHeader
		consistentHashMinimumRingSize = fmt.Sprintf("%v", lb.ConsistentHash.MinimumRingSize)
	} else {
		lbType = "simple"
		lbIdentifier = lb.Simple
	}
	return
}

func getOutlierDetectionVars(od *v1alpha3.OutlierDetection) (httpConsecutiveErrors string, httpInterval string, httpBaseEjectionTime string, httpMaxEjectionPercent string) {
	if od != nil && od.Http != nil {
		httpConsecutiveErrors = fmt.Sprintf("%v", od.Http.ConsecutiveErrors)
		httpInterval = od.Http.BaseEjectionTime.String()
		httpBaseEjectionTime = od.Http.BaseEjectionTime.String()
		httpMaxEjectionPercent = fmt.Sprintf("%v", od.Http.MaxEjectionPercent)
	}
	return
}

func getPortInfo(selector *v1alpha3.PortSelector) (portName string, portNumber string) {
	if selector == nil {
		return
	}

	portName = selector.Name
	portNumber = fmt.Sprintf("%v", selector.Number)
	return
}
