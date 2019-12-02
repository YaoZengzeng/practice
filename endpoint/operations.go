package main

import (
	"bufio"
	"fmt"
	"os"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
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

	endpoint := &apiv1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-endpoint",
			Labels: map[string]string{
				"app": "demo",
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
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, getErr := endpointsClient.Get("demo-endpoint", metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Endpoints: %v", getErr))
		}

		result.ObjectMeta.Labels["app"] = "updateddemo"
		result.ObjectMeta.Annotations["service"] = "updatedabc"
		_, updateErr := endpointsClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Updated endpoint...")

	prompt()
	fmt.Printf("Listing endpoints in namespace %q:\n", apiv1.NamespaceDefault)
	list, err := endpointsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, e := range list.Items {
		fmt.Printf("* %s\n", e.Name)
	}

	// Delete endpoint.
	prompt()
	fmt.Println("Deleting endpoint...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := endpointsClient.Delete("demo-endpoint", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted endpoint.")
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
