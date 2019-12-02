package main

import (
	"bufio"
	"fmt"
	"os"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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

	N := 1000

	for i := 0; i < N; i++ {
		endpoint := &apiv1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("demo-endpoint-%d",i),
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
	}

	// Delete endpoint.
	prompt()

	for i := 0; i < N; i++ {
		fmt.Println("Deleting endpoint...")
		deletePolicy := metav1.DeletePropagationForeground
		if err := endpointsClient.Delete(fmt.Sprintf("demo-endpoint-%d", i), &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}); err != nil {
			panic(err)
		}
		fmt.Println("Deleted endpoint.")
	}
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
