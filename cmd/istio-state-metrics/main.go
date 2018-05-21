package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/douglas-reid/istio-state-metrics/pkg/client/clientset/versioned"
	"github.com/douglas-reid/istio-state-metrics/pkg/options"
)

const (
	metricsPath = "/metrics"
	healthzPath = "/healthz"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

// promLogger implements promhttp.Logger
type promLogger struct{}

func (pl promLogger) Println(v ...interface{}) {
	logger.Print(v)
}

func main() {
	opts := options.NewOptions()
	opts.AddFlags()

	err := opts.Parse()
	if err != nil {
		logger.Fatalf("Error: %s", err)
	}

	if opts.Help {
		opts.Usage()
		os.Exit(0)
	}

	var collectors options.CollectorSet
	if len(opts.Collectors) == 0 {
		logger.Print("Using default collectors")
		collectors = options.DefaultCollectors
	} else {
		collectors = opts.Collectors
	}

	var namespaces options.NamespaceList
	if len(opts.Namespaces) == 0 {
		namespaces = options.DefaultNamespaces
	} else {
		namespaces = opts.Namespaces
	}

	if namespaces.IsAllNamespaces() {
		logger.Print("Using all namespace")
	} else {
		logger.Printf("Using %s namespaces", namespaces)
	}

	kubeClient, err := createKubeClient(opts.Apiserver, opts.Kubeconfig)
	if err != nil {
		logger.Fatalf("Failed to create client: %v", err)
	}

	ksmMetricsRegistry := prometheus.NewRegistry()
	// ksmMetricsRegistry.Register(collectors.ResourcesPerScrapeMetric)
	// ksmMetricsRegistry.Register(collectors.ScrapeErrorTotalMetric)
	ksmMetricsRegistry.Register(prometheus.NewProcessCollector(os.Getpid(), ""))
	ksmMetricsRegistry.Register(prometheus.NewGoCollector())
	go telemetryServer(ksmMetricsRegistry, opts.TelemetryHost, opts.TelemetryPort)

	registry := prometheus.NewRegistry()
	registerCollectors(registry, kubeClient, collectors, namespaces)
	metricsServer(registry, opts.Host, opts.Port)
}

func createKubeClient(apiserver string, kubeconfig string) (versioned.Interface, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	client, err := versioned.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building example clientset: %v", err)
	}

	// Informers don't seem to do a good job logging error messages when it
	// can't reach the server, making debugging hard. This makes it easier to
	// figure out if apiserver is configured incorrectly.
	logger.Printf("Testing communication with server")
	v, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("ERROR communicating with apiserver: %v", err)
	}
	logger.Printf("Running with Kubernetes cluster version: v%s.%s. git version: %s. git tree state: %s. commit: %s. platform: %s",
		v.Major, v.Minor, v.GitVersion, v.GitTreeState, v.GitCommit, v.Platform)
	logger.Printf("Communication with server successful")

	return client, nil
}

func telemetryServer(registry prometheus.Gatherer, host string, port int) {
	// Address to listen on for web interface and telemetry
	listenAddress := net.JoinHostPort(host, strconv.Itoa(port))

	logger.Printf("Starting istio-state-metrics self metrics server: %s", listenAddress)

	mux := http.NewServeMux()

	// Add metricsPath
	mux.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorLog: promLogger{}}))
	// Add index
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Istio-State-Metrics Metrics Server</title></head>
             <body>
             <h1>Istio-State-Metrics Metrics</h1>
			 <ul>
             <li><a href='` + metricsPath + `'>metrics</a></li>
			 </ul>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(listenAddress, mux))
}

func metricsServer(registry prometheus.Gatherer, host string, port int) {
	// Address to listen on for web interface and telemetry
	listenAddress := net.JoinHostPort(host, strconv.Itoa(port))

	logger.Printf("Starting metrics server: %s", listenAddress)

	mux := http.NewServeMux()

	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	// Add metricsPath
	mux.Handle(metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorLog: promLogger{}}))
	// Add healthzPath
	mux.HandleFunc(healthzPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	// Add index
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Istio Metrics Server</title></head>
             <body>
             <h1>Istio Metrics</h1>
			 <ul>
             <li><a href='` + metricsPath + `'>metrics</a></li>
             <li><a href='` + healthzPath + `'>healthz</a></li>
			 </ul>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(listenAddress, mux))
}

// registerCollectors creates and starts informers and initializes and
// registers metrics for collection.
func registerCollectors(registry prometheus.Registerer, client versioned.Interface, enabledCollectors options.CollectorSet, namespaces options.NamespaceList) {
	activeCollectors := []string{}
	for c := range enabledCollectors {
		f, ok := options.AvailableCollectors[c]
		if ok {
			f(registry, client, namespaces)
			activeCollectors = append(activeCollectors, c)
		}
	}

	logger.Printf("Active collectors: %s", strings.Join(activeCollectors, ","))
}
