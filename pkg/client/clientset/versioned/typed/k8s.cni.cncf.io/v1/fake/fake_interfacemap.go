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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	k8scnicncfiov1 "github.com/dougbtv/surveyor-cni/pkg/apis/k8s.cni.cncf.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeInterfaceMaps implements InterfaceMapInterface
type FakeInterfaceMaps struct {
	Fake *FakeK8sV1
	ns   string
}

var interfacemapsResource = schema.GroupVersionResource{Group: "k8s.cni.cncf.io", Version: "v1", Resource: "interfacemaps"}

var interfacemapsKind = schema.GroupVersionKind{Group: "k8s.cni.cncf.io", Version: "v1", Kind: "InterfaceMap"}

// Get takes name of the interfaceMap, and returns the corresponding interfaceMap object, and an error if there is any.
func (c *FakeInterfaceMaps) Get(ctx context.Context, name string, options v1.GetOptions) (result *k8scnicncfiov1.InterfaceMap, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(interfacemapsResource, c.ns, name), &k8scnicncfiov1.InterfaceMap{})

	if obj == nil {
		return nil, err
	}
	return obj.(*k8scnicncfiov1.InterfaceMap), err
}

// List takes label and field selectors, and returns the list of InterfaceMaps that match those selectors.
func (c *FakeInterfaceMaps) List(ctx context.Context, opts v1.ListOptions) (result *k8scnicncfiov1.InterfaceMapList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(interfacemapsResource, interfacemapsKind, c.ns, opts), &k8scnicncfiov1.InterfaceMapList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &k8scnicncfiov1.InterfaceMapList{ListMeta: obj.(*k8scnicncfiov1.InterfaceMapList).ListMeta}
	for _, item := range obj.(*k8scnicncfiov1.InterfaceMapList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested interfaceMaps.
func (c *FakeInterfaceMaps) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(interfacemapsResource, c.ns, opts))

}

// Create takes the representation of a interfaceMap and creates it.  Returns the server's representation of the interfaceMap, and an error, if there is any.
func (c *FakeInterfaceMaps) Create(ctx context.Context, interfaceMap *k8scnicncfiov1.InterfaceMap, opts v1.CreateOptions) (result *k8scnicncfiov1.InterfaceMap, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(interfacemapsResource, c.ns, interfaceMap), &k8scnicncfiov1.InterfaceMap{})

	if obj == nil {
		return nil, err
	}
	return obj.(*k8scnicncfiov1.InterfaceMap), err
}

// Update takes the representation of a interfaceMap and updates it. Returns the server's representation of the interfaceMap, and an error, if there is any.
func (c *FakeInterfaceMaps) Update(ctx context.Context, interfaceMap *k8scnicncfiov1.InterfaceMap, opts v1.UpdateOptions) (result *k8scnicncfiov1.InterfaceMap, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(interfacemapsResource, c.ns, interfaceMap), &k8scnicncfiov1.InterfaceMap{})

	if obj == nil {
		return nil, err
	}
	return obj.(*k8scnicncfiov1.InterfaceMap), err
}

// Delete takes name of the interfaceMap and deletes it. Returns an error if one occurs.
func (c *FakeInterfaceMaps) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(interfacemapsResource, c.ns, name, opts), &k8scnicncfiov1.InterfaceMap{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeInterfaceMaps) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(interfacemapsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &k8scnicncfiov1.InterfaceMapList{})
	return err
}

// Patch applies the patch and returns the patched interfaceMap.
func (c *FakeInterfaceMaps) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *k8scnicncfiov1.InterfaceMap, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(interfacemapsResource, c.ns, name, pt, data, subresources...), &k8scnicncfiov1.InterfaceMap{})

	if obj == nil {
		return nil, err
	}
	return obj.(*k8scnicncfiov1.InterfaceMap), err
}
