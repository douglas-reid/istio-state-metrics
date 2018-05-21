package options

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

type options struct {
	Apiserver     string
	Kubeconfig    string
	Help          bool
	Port          int
	Host          string
	TelemetryPort int
	TelemetryHost string
	Collectors    CollectorSet
	Namespaces    NamespaceList
	Version       bool

	flags *pflag.FlagSet
}

func NewOptions() *options {
	return &options{
		Collectors: CollectorSet{},
	}
}

func (o *options) AddFlags() {
	o.flags = pflag.NewFlagSet("", pflag.ExitOnError)

	o.flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		o.flags.PrintDefaults()
	}

	o.flags.StringVar(&o.Apiserver, "apiserver", "", `The URL of the apiserver to use as a master`)
	o.flags.StringVar(&o.Kubeconfig, "kubeconfig", "", "Absolute path to the kubeconfig file")
	o.flags.BoolVarP(&o.Help, "help", "h", false, "Print Help text")
	o.flags.IntVar(&o.Port, "port", 9090, `Port to expose metrics on.`)
	o.flags.StringVar(&o.Host, "host", "0.0.0.0", `Host to expose metrics on.`)
	o.flags.IntVar(&o.TelemetryPort, "telemetry-port", 9093, `Port to expose istio-state-metrics self metrics on.`)
	o.flags.StringVar(&o.TelemetryHost, "telemetry-host", "0.0.0.0", `Host to expose istio-state-metrics self metrics on.`)
	o.flags.Var(&o.Collectors, "collectors", fmt.Sprintf("Comma-separated list of collectors to be enabled. Defaults to %q", &DefaultCollectors))
	o.flags.Var(&o.Namespaces, "namespace", fmt.Sprintf("Comma-separated list of namespaces to be enabled. Defaults to %q", &DefaultNamespaces))
}

func (o *options) Parse() error {
	err := o.flags.Parse(os.Args)
	return err
}

func (o *options) Usage() {
	o.flags.Usage()
}
