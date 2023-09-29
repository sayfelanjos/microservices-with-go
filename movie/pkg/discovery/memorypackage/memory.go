package memory

import (
	"context"
	"errors"
	"movieexample.com/movie/pkg/discovery"
	"sync"
	"time"
)

type ServiceName string
type InstanceID string

// Registry defines an in-memory service registry.
type Registry struct {
	sync.RWMutex
	serviceAddrs map[ServiceName]map[InstanceID]*serviceInstance
}
type serviceInstance struct {
	hostPort   string
	lastActive time.Time
}

// NewRegistry creates a new in-memory service
// registry instance.
func NewRegistry() *Registry {
	return &Registry{serviceAddrs: map[ServiceName]map[InstanceID]*serviceInstance{}}
}

// Register creates a service record in the registry.
func (r *Registry) Register(ctx context.Context, instanceID InstanceID, serviceName ServiceName, hostPort string) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[serviceName]; !ok {
		r.serviceAddrs[serviceName] = map[InstanceID]*serviceInstance{}
	}
	r.serviceAddrs[serviceName][instanceID] = &serviceInstance{hostPort: hostPort, lastActive: time.Now()}
	return nil
}

// Deregister removes a service record from the
// registry.
func (r *Registry) Deregister(ctx context.Context, instanceID InstanceID, serviceName ServiceName) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[serviceName]; !ok {
		return nil
	}
	delete(r.serviceAddrs[serviceName], instanceID)
	return nil
}

// ReportHealthyState is a push mechanism for
// reporting healthy state to the registry.
func (r *Registry) ReportHealthyState(instanceID InstanceID, serviceName ServiceName) error {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.serviceAddrs[serviceName]; !ok {
		return errors.New("service is not registered yet")
	}
	if _, ok := r.serviceAddrs[serviceName][instanceID]; !ok {
		return errors.New("service instance is not registered yet")
	}
	r.serviceAddrs[serviceName][instanceID].lastActive = time.Now()
	return nil
}

// ServiceAddresses returns the list of addresses of
// active instances of the given service.
func (r *Registry) ServiceAddresses(ctx context.Context, serviceName ServiceName) ([]string, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.serviceAddrs[serviceName]) == 0 {
		return nil, discovery.ErrNotFound
	}
	var res []string
	for _, i := range r.serviceAddrs[serviceName] {
		if i.lastActive.Before(time.Now().Add(-5 * time.Second)) {
			continue
		}
		res = append(res, i.hostPort)
	}
	return res, nil
}
