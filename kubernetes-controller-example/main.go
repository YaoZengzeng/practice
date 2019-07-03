package main

import (
	"flag"
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	examplev1 "github.com/YaoZengzeng/practice/kubernetes-controller-example/pkg/apis/example/v1"
	clientset "github.com/YaoZengzeng/practice/kubernetes-controller-example/pkg/client/clientset/versioned"

	"k8s.io/client-go/util/workqueue"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	resyncPeriod = 5 * time.Minute
)

var (
	master string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "/root/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster")
	flag.StringVar(&master, "master", "", "The address of the Kubernetes API Server. Overrides any value in kubeconfig. Only required if out-of-cluster")
}

type Controller struct {
	kclient 	kubernetes.Interface
	eclient 	clientset.Interface

	examInf		cache.SharedIndexInformer

	queue		workqueue.RateLimitingInterface
}

func main() {
	flag.Parse()
	// Loading kubernetes config.
	cfg, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v", err)
		return
	}

	// Building kubernetes client.
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("Error building kubernetes clientset: %v", err)
		return
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("Error building example client: %v", err)
		return
	}

	c := &Controller{
		kclient:	kubeClient,
		eclient:	exampleClient,

		queue:		workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "example"),
	}

	c.examInf = cache.NewSharedIndexInformer(&cache.ListWatch{
			ListFunc:	func(options metav1.ListOptions) (runtime.Object, error) {
				return exampleClient.ExampleV1().Dogs(v1.NamespaceAll).List(options)
			},
			WatchFunc:	exampleClient.ExampleV1().Dogs(v1.NamespaceAll).Watch,
		},
		&examplev1.Dog{}, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	stopc := make(chan struct{})

	go c.examInf.Run(stopc)

	if !cache.WaitForCacheSync(stopc, c.examInf.HasSynced) {
		fmt.Printf("Failed to sync Dog cache")
		return
	}

	c.examInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:	c.handleDogAdd,
		UpdateFunc:	c.handleDogUpdate,
		DeleteFunc:	c.handleDogDelete,
	})

	// Never stop actually.
	<-stopc
}

func (c *Controller) handleDogAdd(obj interface{}) {
	dog := obj.(*examplev1.Dog)
	fmt.Printf("Handle Dog Add: %v\n", dog)

}

func (c *Controller) handleDogUpdate(old, cur interface{}) {
	if old.(*examplev1.Dog).ResourceVersion == cur.(*examplev1.Dog).ResourceVersion {
		fmt.Printf("Handle Dog Update, ResourceVersion equal: %v, skip", old.(*examplev1.Dog).ResourceVersion)
		return
	}

	dog := cur.(*examplev1.Dog)
	fmt.Printf("Handle Dog Update: %v\n", dog)
}

func (c *Controller) handleDogDelete(obj interface{}) {
	fmt.Printf("Handle Dog Delete: %v\n", obj)
}
