package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {
	os.Setenv("KUBERNETES_SERVICE_HOST", "10.247.0.1")
	os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	endpointsClient := clientset.CoreV1().Endpoints(apiv1.NamespaceDefault)

	endpoint := &apiv1.Endpoint{
		ObjectMeta:	metav1.ObjectMeta{
			Name:	"demo-endpoint",
			Labels:	map[string]string{
				"app":	"demo",
			},
			Annotations: map[string]string{
				"service": "abc",
			},
		},
	}

	// Create endpoint
	fmt.Printf("Creating endpoint...")
	result, err := endpointsClient.Create(endpoint)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created endpoint %q.\n", result.GetObjectMeta().GetName())

	// Update endpoint.
	prompt()
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
