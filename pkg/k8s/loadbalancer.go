package k8s

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EnsureLBService creates the required LoadBalancer if it does not exist
func (m *Manager) EnsureLBService() error {
	_, err := m.client.CoreV1().Services(m.serviceNamespace).Get(m.serviceName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = m.client.CoreV1().Services(m.serviceNamespace).Create(&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: m.serviceName,
			},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{getServicePortForTargetService(keepaliveTargetService)},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create loadbalancer service: %v", err)
		}

	}

	return err
}

func (m *Manager) updateLBService(targetServices []targetService) error {
	updated := false
	service, err := m.serviceIndexer.Services(m.serviceNamespace).Get(m.serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %v", err)
	}

	//Add missing ports
	for _, ts := range targetServices {
		if !targetServiceExistsInLoadBalancer(service, ts) {
			updated = true
			service.Spec.Ports = append(service.Spec.Ports, getServicePortForTargetService(ts))
		}
	}

	//Delete old ports
	for i, sp := range service.Spec.Ports {
		if !servicePortExistsInTargetServiceSlice(sp, targetServices) {
			updated = true
			service.Spec.Ports = append(service.Spec.Ports[:i], service.Spec.Ports[i+1:]...)
		}
	}

	if updated {
		_, err = m.client.CoreV1().Services(m.serviceNamespace).Update(service)
		return err
	}
	return nil
}

func getServicePortForTargetService(ts targetService) corev1.ServicePort {
	return corev1.ServicePort{
		Port:       ts.nodePort,
		TargetPort: intstr.IntOrString{IntVal: ts.nodePort, Type: intstr.Int},
		Protocol:   corev1.ProtocolTCP,
		Name:       ts.name,
	}
}

func targetServiceExistsInLoadBalancer(service *corev1.Service, ts targetService) bool {
	for _, p := range service.Spec.Ports {
		if p.Port == ts.nodePort && p.Name == ts.name {
			return true
		}
	}
	return false
}

func servicePortExistsInTargetServiceSlice(servicePort corev1.ServicePort, targetServices []targetService) bool {
	for _, p := range targetServices {
		if p.nodePort == servicePort.Port && p.name == servicePort.Name {
			return true
		}
	}
	return false
}
