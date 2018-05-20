package main

import (
	"flag"
	"fmt"
	"log"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/douglas-reid/istio-state-metrics/pkg/client/clientset/versioned"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	master     = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building example clientset: %v", err)
	}

	list, err := exampleClient.ConfigV1alpha2().Rules("").List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing all rules: %v", err)
	}

	for _, rule := range list.Items {
		fmt.Printf("rule %s with match %q\n", rule.Name, rule.Spec.Match)
	}
}
