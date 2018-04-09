package k8s

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureEndpoint creates the required endpoint if it does not exist
func (m *Manager) EnsureEndpoint() error {
	_, err := m.client.CoreV1().Endpoints(m.serviceNamespace).Get(m.serviceName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = m.client.CoreV1().Endpoints(m.serviceNamespace).Create(&corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name: m.serviceName,
			},
			Subsets: []corev1.EndpointSubset{getEndpointSubsetForTargetService(keepaliveTargetService)},
		})
		if err != nil {
			return fmt.Errorf("failed to create endpoint: %v", err)
		}
	}

	return err
}

func getEndpointSubsetForTargetService(ts targetService) corev1.EndpointSubset {
	return corev1.EndpointSubset{
		Ports:     []corev1.EndpointPort{{Port: ts.nodePort, Protocol: ts.protocol}},
		Addresses: []corev1.EndpointAddress{{IP: ts.clusterIP}},
	}
}

func (m *Manager) updateEndpoints(targetServices []targetService) error {
	updated := false
	endpoint, err := m.endpointIndexer.Endpoints(m.serviceNamespace).Get(m.serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %v", err)
	}

	//Add missing ports to endpoint
	for _, ts := range targetServices {
		if !targetServiceExistsInEndpoint(endpoint, ts) {
			updated = true
			endpoint.Subsets = append(endpoint.Subsets, getEndpointSubsetForTargetService(ts))
		}
	}

	//Delete old ports
	updatedSubsets := endpoint.Subsets[:0]
	for _, sub := range endpoint.Subsets {
		if !subsetExistsInTargetServiceSlice(sub, targetServices) {
			updated = true
			continue
		}
		updatedSubsets = append(updatedSubsets, sub)
	}
	endpoint.Subsets = updatedSubsets

	if updated {
		_, err = m.client.CoreV1().Endpoints(m.serviceNamespace).Update(endpoint)
		return err
	}
	return nil
}

func targetServiceExistsInEndpoint(endpoint *corev1.Endpoints, ts targetService) bool {
	for _, s := range endpoint.Subsets {
		if s.Addresses[0].IP == ts.clusterIP && s.Ports[0].Port == ts.nodePort {
			return true
		}
	}
	return false
}

func subsetExistsInTargetServiceSlice(subset corev1.EndpointSubset, targetServices []targetService) bool {
	for _, p := range targetServices {
		if p.nodePort == subset.Ports[0].Port && p.clusterIP == subset.Addresses[0].IP {
			return true
		}
	}
	return false
}
