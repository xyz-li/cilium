// Copyright 2017-2020 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	corev1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePods implements PodInterface
type FakePods struct {
	Fake *FakeCoreV1
	ns   string
}

var podsResource = schema.GroupVersionResource{Group: "core", Version: "v1", Resource: "pods"}

var podsKind = schema.GroupVersionKind{Group: "core", Version: "v1", Kind: "Pod"}

// Get takes name of the pod, and returns the corresponding pod object, and an error if there is any.
func (c *FakePods) Get(ctx context.Context, name string, options v1.GetOptions) (result *corev1.Pod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(podsResource, c.ns, name), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// List takes label and field selectors, and returns the list of Pods that match those selectors.
func (c *FakePods) List(ctx context.Context, opts v1.ListOptions) (result *corev1.PodList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(podsResource, podsKind, c.ns, opts), &corev1.PodList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &corev1.PodList{ListMeta: obj.(*corev1.PodList).ListMeta}
	for _, item := range obj.(*corev1.PodList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested pods.
func (c *FakePods) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(podsResource, c.ns, opts))

}

// Create takes the representation of a pod and creates it.  Returns the server's representation of the pod, and an error, if there is any.
func (c *FakePods) Create(ctx context.Context, pod *corev1.Pod, opts v1.CreateOptions) (result *corev1.Pod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(podsResource, c.ns, pod), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// Update takes the representation of a pod and updates it. Returns the server's representation of the pod, and an error, if there is any.
func (c *FakePods) Update(ctx context.Context, pod *corev1.Pod, opts v1.UpdateOptions) (result *corev1.Pod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(podsResource, c.ns, pod), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePods) UpdateStatus(ctx context.Context, pod *corev1.Pod, opts v1.UpdateOptions) (*corev1.Pod, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(podsResource, "status", c.ns, pod), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// Delete takes name of the pod and deletes it. Returns an error if one occurs.
func (c *FakePods) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(podsResource, c.ns, name), &corev1.Pod{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePods) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(podsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &corev1.PodList{})
	return err
}

// Patch applies the patch and returns the patched pod.
func (c *FakePods) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *corev1.Pod, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podsResource, c.ns, name, pt, data, subresources...), &corev1.Pod{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.Pod), err
}

// GetEphemeralContainers takes name of the pod, and returns the corresponding ephemeralContainers object, and an error if there is any.
func (c *FakePods) GetEphemeralContainers(ctx context.Context, podName string, options v1.GetOptions) (result *corev1.EphemeralContainers, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetSubresourceAction(podsResource, c.ns, "ephemeralcontainers", podName), &corev1.EphemeralContainers{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.EphemeralContainers), err
}

// UpdateEphemeralContainers takes the representation of a ephemeralContainers and updates it. Returns the server's representation of the ephemeralContainers, and an error, if there is any.
func (c *FakePods) UpdateEphemeralContainers(ctx context.Context, podName string, ephemeralContainers *corev1.EphemeralContainers, opts v1.UpdateOptions) (result *corev1.EphemeralContainers, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(podsResource, "ephemeralcontainers", c.ns, ephemeralContainers), &corev1.EphemeralContainers{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.EphemeralContainers), err
}
