package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/kubermatic/nodeport-exposer/pkg/controller"
	"github.com/kubermatic/nodeport-exposer/pkg/k8s"

	coreinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig string
	var master string
	var lbService string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&lbService, "lb-service-name", "nodeport-exposer/nodeport-lb", "name of the load-balancer service to manage. If it does not exist, a new one will be created. Format: 'ns/name' (Default: nodeport-exposer/nodeport-lb)")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		glog.Fatal(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}

	coreInformers := coreinformers.NewSharedInformerFactory(client, 5*time.Minute)
	serviceInformer := coreInformers.Core().V1().Services()
	endpointInformer := coreInformers.Core().V1().Endpoints()

	ns, name, err := cache.SplitMetaNamespaceKey(lbService)
	if err != nil {
		glog.Fatalf("invalid value for -lb-service-name : %v", err)
	}
	manager := k8s.NewManager(client, serviceInformer.Lister(), endpointInformer.Lister(), ns, name)

	glog.V(6).Infof("ensuring that load balancer service exists")
	if err := manager.EnsureLBService(); err != nil {
		glog.Fatalf("failed to ensure that the lb service exists: %v", err)
	}

	glog.V(6).Infof("ensuring that load balancer service endpoints exists")
	if err := manager.EnsureEndpoint(); err != nil {
		glog.Fatalf("failed to ensure that the lb service endpoints exists: %v", err)
	}

	c := controller.NewController(serviceInformer, manager)

	stop := make(chan struct{})
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		close(stop)
	}()

	go serviceInformer.Informer().Run(stop)
	go endpointInformer.Informer().Run(stop)

	c.Run(stop)
}
