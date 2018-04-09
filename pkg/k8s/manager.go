package k8s

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
)

// Manager exposes NodePorts via a LoadBalancer+Endpoints
type Manager struct {
	serviceNamespace string
	serviceName      string
	client           kubernetes.Interface
	serviceIndexer   listerscorev1.ServiceLister
	endpointIndexer  listerscorev1.EndpointsLister
}

type targetService struct {
	clusterIP string
	nodePort  int32
	name      string
	protocol  corev1.Protocol
}

var keepaliveTargetService = targetService{
	clusterIP: "10.47.240.1",
	nodePort:  30000,
	name:      "keepalive",
}

// NewManager returns a Manager which exposes NodePorts via a LoadBalancer+Endpoints
func NewManager(client kubernetes.Interface, serviceIndexer listerscorev1.ServiceLister, endpointIndexer listerscorev1.EndpointsLister, serviceNamespace, serviceName string) *Manager {
	return &Manager{
		client:           client,
		serviceName:      serviceName,
		serviceNamespace: serviceNamespace,
		serviceIndexer:   serviceIndexer,
		endpointIndexer:  endpointIndexer,
	}
}

// Update exposes all nodeports from the given services via LoadBalancer+Endpoints
func (m *Manager) Update(services []*corev1.Service) error {
	var want []targetService
	//Get wanted list
	for _, s := range services {
		for _, sp := range s.Spec.Ports {
			if sp.NodePort != 0 {
				want = append(want, targetService{
					clusterIP: s.Spec.ClusterIP,
					nodePort:  sp.NodePort,
					name:      fmt.Sprintf("%s-%s-%d", s.Namespace, s.Name, sp.NodePort),
					protocol:  sp.Protocol,
				})
			}
		}
	}
	if len(want) == 0 {
		want = append(want, keepaliveTargetService)
	}

	if err := m.updateLBService(want); err != nil {
		return err
	}
	if err := m.updateEndpoints(want); err != nil {
		return err
	}
	return nil
}
