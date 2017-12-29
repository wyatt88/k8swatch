package controller

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/wyatt88/k8swatch/pkg/handlers"
	"github.com/wyatt88/k8swatch/pkg/utils"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller is Ok
type Controller struct {
	clientset    kubernetes.Interface
	indexer      cache.Indexer
	queue        workqueue.RateLimitingInterface
	informer     cache.Controller
	eventHandler handlers.AlertManager
}

// Start is Ok
func Start(kubeConfig string, master string, eventHandler handlers.AlertManager) {
	var kubeClient kubernetes.Interface
	if _, err := os.Stat(kubeConfig); os.IsNotExist(err) {
		kubeClient = utils.GetClientOutOfCluster(kubeConfig, master)
		glog.Errorf("Kubeconfig file doesn't exist;Error is %v", err)
	} else {
		kubeClient = utils.GetClient()
	}

	c := newController(kubeClient, eventHandler)
	stopCh := make(chan struct{})
	defer close(stopCh)
	go c.Run(stopCh)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm

}

func newController(client kubernetes.Interface, eventHandler handlers.AlertManager) *Controller {
	eventListWatcher := cache.NewListWatchFromClient(client.CoreV1().RESTClient(), "events", v1.NamespaceAll, fields.Everything())
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	indexer, informer := cache.NewIndexerInformer(eventListWatcher, &v1.Event{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})
	return &Controller{
		clientset:    client,
		informer:     informer,
		queue:        queue,
		eventHandler: eventHandler,
		indexer:      indexer,
	}
}

// Run is ok
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	glog.Info("Starting k8swatch controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	glog.Info("k8swatch controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)
	err := c.processItem(key.(string))
	c.handleErr(err, key)
	return true
}

func (c *Controller) processItem(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching obj with key %s failed with %v", key, err)
		return err
	}
	if !exists {
		fmt.Printf("Event %s does not exist anymore \n", key)
	}
	c.eventHandler.ObjectCreated(obj)
	return nil
}

func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	} else if c.queue.NumRequeues(key) < 5 {
		glog.Infof("Error syncing pod %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	} else {
		c.queue.Forget(key)
		runtime.HandleError(err)
		glog.Infof("Dropping pod %q out of the queue: %v", key, err)
	}
}
