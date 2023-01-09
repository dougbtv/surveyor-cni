/*
Copyright 2023 The Kubernetes Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/dougbtv/surveyor-cni/pkg/apis/k8s.cni.cncf.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// InterfaceMapLister helps list InterfaceMaps.
// All objects returned here must be treated as read-only.
type InterfaceMapLister interface {
	// List lists all InterfaceMaps in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.InterfaceMap, err error)
	// InterfaceMaps returns an object that can list and get InterfaceMaps.
	InterfaceMaps(namespace string) InterfaceMapNamespaceLister
	InterfaceMapListerExpansion
}

// interfaceMapLister implements the InterfaceMapLister interface.
type interfaceMapLister struct {
	indexer cache.Indexer
}

// NewInterfaceMapLister returns a new InterfaceMapLister.
func NewInterfaceMapLister(indexer cache.Indexer) InterfaceMapLister {
	return &interfaceMapLister{indexer: indexer}
}

// List lists all InterfaceMaps in the indexer.
func (s *interfaceMapLister) List(selector labels.Selector) (ret []*v1.InterfaceMap, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.InterfaceMap))
	})
	return ret, err
}

// InterfaceMaps returns an object that can list and get InterfaceMaps.
func (s *interfaceMapLister) InterfaceMaps(namespace string) InterfaceMapNamespaceLister {
	return interfaceMapNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// InterfaceMapNamespaceLister helps list and get InterfaceMaps.
// All objects returned here must be treated as read-only.
type InterfaceMapNamespaceLister interface {
	// List lists all InterfaceMaps in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.InterfaceMap, err error)
	// Get retrieves the InterfaceMap from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.InterfaceMap, error)
	InterfaceMapNamespaceListerExpansion
}

// interfaceMapNamespaceLister implements the InterfaceMapNamespaceLister
// interface.
type interfaceMapNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all InterfaceMaps in the indexer for a given namespace.
func (s interfaceMapNamespaceLister) List(selector labels.Selector) (ret []*v1.InterfaceMap, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.InterfaceMap))
	})
	return ret, err
}

// Get retrieves the InterfaceMap from the indexer for a given namespace and name.
func (s interfaceMapNamespaceLister) Get(name string) (*v1.InterfaceMap, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("interfacemap"), name)
	}
	return obj.(*v1.InterfaceMap), nil
}
