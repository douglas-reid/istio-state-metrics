package collectors

import (
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
)

var (
	resyncPeriod       = 5 * time.Minute
	invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

	ScrapeErrorTotalMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "scrape_error_total",
			Namespace: "istio_state_metrics",
			Help:      "Total scrape errors encountered when scraping a resource",
		},
		[]string{"resource"},
	)

	ResourcesPerScrapeMetric = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:      "resources_per_scrape",
			Namespace: "istio_state_metrics",
			Help:      "Number of resources returned per scrape",
		},
		[]string{"resource"},
	)
)

type SharedInformerList []cache.SharedInformer

type NewInformerFn func(namespace string) cache.SharedIndexInformer

func NewSharedInformerList(fn NewInformerFn, namespaces []string) *SharedInformerList {
	sinfs := SharedInformerList{}
	for _, namespace := range namespaces {
		sinfs = append(sinfs, fn(namespace))
	}
	return &sinfs
}

func (sil SharedInformerList) Run(stopCh <-chan struct{}) {
	for _, sinf := range sil {
		go sinf.Run(stopCh)
	}
}

func promSafeName(in string) string {
	return invalidLabelCharRE.ReplaceAllString(in, "_")
}
