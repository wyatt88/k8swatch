package controller

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/util/wait"
	"github.com/wyatt88/k8swatch/pkg/handlers"
	"github.com/wyatt88/k8swatch/pkg/utils"
	"os"
	"os/signal"
	"syscall"
	"k8s.io/apimachinery/pkg/fields"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/runtime"
	"fmt"
	"time"
)

type Controller struct {
	clientset kubernetes.Interface
	indexer cache.Indexer
	queue workqueue.RateLimitingInterface
	informer cache.Controller
	eventHandler handlers.AlertManager
}

func Start(master string,eventHandler handlers.AlertManager)  {
	kubeClient := utils.GetClientOutOfCluster()
	if master != "" {
		c := newController(kubeClient,eventHandler)
		stopCh := make(chan struct{})
		defer close(stopCh)
		go c.Run(stopCh)
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGTERM)
		signal.Notify(sigterm, syscall.SIGINT)
		<-sigterm
	}
}

func newController(client kubernetes.Interface,eventHandler handlers.AlertManager) *Controller {
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
		indexer: indexer,
	}
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()
	
	glog.Info("Starting kubewatch controller")
	
	go c.informer.Run(stopCh)
	
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	
	glog.Info("Kubewatch controller synced and ready")
	
	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	key,quit := c.queue.Get()
	if quit{
		return false
	}
	defer c.queue.Done(key)
	err := c.processItem(key.(string))
	c.handleErr(err,key)
	return true
}

func (c *Controller) processItem(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching obj with key %s failed with %v",key,err)
		return err
	}
	if !exists {
		fmt.Printf("Event %s does not exist anymore \n",key)
	}
	c.eventHandler.ObjectCreated(obj)
	return nil
}

func (c *Controller) handleErr(err error,key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	} else if c.queue.NumRequeues(key) < 5{
		glog.Infof("Error syncing pod %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	} else {
		c.queue.Forget(key)
		runtime.HandleError(err)
		glog.Infof("Dropping pod %q out of the queue: %v", key, err)
	}
}