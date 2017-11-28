package controller

import (
	"errors"
	"fmt"

	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	coreinformersv1 "k8s.io/client-go/informers/core/v1"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	exposeAnnotationKey = "nodeport-exposer.k8s.io/expose"
)

// NodePortExposer is a controller which exposes NodePorts
type NodePortExposer struct {
	serviceIndexer  listerscorev1.ServiceLister
	serviceInformer cache.SharedIndexInformer
	em              ExposeManager
}

// ExposeManager manages the actual exposing of NodePorts
type ExposeManager interface {
	Update(services []*corev1.Service) error
}

// NewController returns a new instance of the NodePortExposer controller
func NewController(serviceInformer coreinformersv1.ServiceInformer, em ExposeManager) *NodePortExposer {
	return &NodePortExposer{
		serviceIndexer:  serviceInformer.Lister(),
		serviceInformer: serviceInformer.Informer(),
		em:              em,
	}
}

func (c *NodePortExposer) syncExposeManager() error {
	services, err := c.serviceIndexer.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to list services: %v", err)
	}

	var exposeList []*corev1.Service
	for _, s := range services {
		if s.Annotations[exposeAnnotationKey] == "true" {
			glog.V(6).Infof("exposing nodeports from service %s/%s", s.Namespace, s.Name)
			exposeList = append(exposeList, s)
		} else {
			glog.V(6).Infof("ignoring service %s/%s as the annotation %s=true is missing", s.Namespace, s.Name, exposeAnnotationKey)
		}
	}

	err = c.em.Update(exposeList)
	if err != nil {
		return fmt.Errorf("failed to update expose manager: %v", err)
	}

	return nil
}

// Run starts the control loop. This function is blocking
func (c *NodePortExposer) Run(stopCh chan struct{}) {
	defer runtime.HandleCrash()

	glog.V(4).Infof("Starting Service controller")

	if !cache.WaitForCacheSync(stopCh, c.serviceInformer.HasSynced) {
		runtime.HandleError(errors.New("timed out waiting for caches to sync"))
		return
	}

	c.serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if err := c.syncExposeManager(); err != nil {
				glog.Errorf("failed to sync: %v", err)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			if err := c.syncExposeManager(); err != nil {
				glog.Errorf("failed to sync: %v", err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if err := c.syncExposeManager(); err != nil {
				glog.Errorf("failed to sync: %v", err)
			}
		},
	})

	if err := c.syncExposeManager(); err != nil {
		glog.Errorf("failed to sync: %v", err)
	}
	<-stopCh
	glog.V(4).Infof("Stopping Service controller")
}
