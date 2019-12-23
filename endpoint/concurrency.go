package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func main() {
	os.Setenv("KUBERNETES_SERVICE_HOST", "10.247.0.1")
	os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	informers := informers.NewSharedInformerFactory(clientset, 0)
	client := clientset.CoreV1().Endpoints(apiv1.NamespaceDefault)
	lister := informers.Core().V1().Endpoints().Lister().Endpoints(apiv1.NamespaceDefault)

	endpointInformer := informers.Core().V1().Endpoints().Informer()

	stop := make(chan struct{})
	informers.Start(stop)

	if !cache.WaitForCacheSync(stop, endpointInformer.HasSynced) {
		panic("failed to wait endpoint informer synced")
	}

	name := "concurrency-endpoint-test"
	endpoint := &apiv1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	var errnum int32

	// Launch N goroutines to create endpoint concurrently.
	N := 100
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			if _, err := client.Create(endpoint); err != nil {
				if errors.IsAlreadyExists(err) {
					atomic.AddInt32(&errnum, 1)
					log.Printf("create endpoints failed, the error is IsAlreadyExists()")
					return
				}
				log.Printf("create endpoints faied: %v\n", err)
			}
		}()
	}
	wg.Wait()
	log.Printf("N is %v, number of IsAlreadyExits is %v", N, errnum)

	errnum = 0
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(n int) {
			defer wg.Done()
			result, err := lister.Get(name)
			if err != nil {
				log.Printf("get endpoint %v from cache failed: %v", name, err)
				return
			}
			tmp := result.DeepCopy()
			if tmp.Annotations == nil {
				tmp.Annotations = make(map[string]string)
			}
			tmp.Annotations["update-added-key"] = fmt.Sprintf("index-%d", n)
			if _, err := client.Update(tmp); err != nil {
				if errors.IsConflict(err) {
					atomic.AddInt32(&errnum, 1)
					log.Printf("update endpoints failed, error is IsConflict")
					return
				}
				log.Printf("update endpoints failed: %v", err)
			}
		}(i)
	}
	wg.Wait()
	log.Printf("N is %v, errnum is %v", N, errnum)

	obj, err := lister.Get(name)
	if err != nil {
		log.Printf("get endpoint %v from cache failed: %v", name, err)
		return
	}
	tmp := obj.DeepCopy()
	old := obj.DeepCopy()
	if tmp.Annotations == nil {
		tmp.Annotations = make(map[string]string)
	}
	tmp.Annotations["update-final-key"] = "final"
	if _, err := client.Update(tmp); err != nil {
		if errors.IsConflict(err) {
			log.Printf("update final endpoints failed, err is Confilect")
			return
		}
		log.Printf("update endpoints failed: %v", err)
	}

	for {
		ne, err := lister.Get(name)
		if err != nil {
			log.Printf("get endpoint %v from cache failed: %v", name, err)
			return
		}
		if reflect.DeepEqual(old, ne) {
			log.Printf("cache has not been updated, sleep 100 * time.Microsecond")
			time.Sleep(100 * time.Microsecond)
		} else {
			log.Printf("cache has been updated")
			break
		}
	}

	deletePolicy := metav1.DeletePropagationForeground
	if err := client.Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Printf("delete endpoint failed: %v\n", err)
	}
	log.Printf("finally, %v get deleted", name)
}
