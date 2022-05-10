/*
Copyright 2018 The Kubernetes Authors.

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

package v1alpha1

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE k8s.restdev.com/operators/pkg/client/clientset/versioned/typed/scaling/v1alpha1 ScheduledScalersGetter,ScheduledScalerInterface

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "k8s.restdev.com/operators/pkg/apis/scaling/v1alpha1"
	scheme "k8s.restdev.com/operators/pkg/client/clientset/versioned/scheme"
)

// ScheduledScalersGetter has a method to return a ScheduledScalerInterface.
// A group's client should implement this interface.
type ScheduledScalersGetter interface {
	ScheduledScalers(namespace string) ScheduledScalerInterface
}

// ScheduledScalerInterface has methods to work with ScheduledScaler resources.
type ScheduledScalerInterface interface {
	Create(*v1alpha1.ScheduledScaler) (*v1alpha1.ScheduledScaler, error)
	Update(*v1alpha1.ScheduledScaler) (*v1alpha1.ScheduledScaler, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ScheduledScaler, error)
	List(opts v1.ListOptions) (*v1alpha1.ScheduledScalerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ScheduledScaler, err error)
	ScheduledScalerExpansion
}

// scheduledScalers implements ScheduledScalerInterface
type scheduledScalers struct {
	client rest.Interface
	ns     string
}

// newScheduledScalers returns a ScheduledScalers
func newScheduledScalers(c *ScalingV1alpha1Client, namespace string) *scheduledScalers {
	return &scheduledScalers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the scheduledScaler, and returns the corresponding scheduledScaler object, and an error if there is any.
func (c *scheduledScalers) Get(name string, options v1.GetOptions) (result *v1alpha1.ScheduledScaler, err error) {
	result = &v1alpha1.ScheduledScaler{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("scheduledscalers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ScheduledScalers that match those selectors.
func (c *scheduledScalers) List(opts v1.ListOptions) (result *v1alpha1.ScheduledScalerList, err error) {
	result = &v1alpha1.ScheduledScalerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("scheduledscalers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested scheduledScalers.
func (c *scheduledScalers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("scheduledscalers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(context.TODO())
}

// Create takes the representation of a scheduledScaler and creates it.  Returns the server's representation of the scheduledScaler, and an error, if there is any.
func (c *scheduledScalers) Create(scheduledScaler *v1alpha1.ScheduledScaler) (result *v1alpha1.ScheduledScaler, err error) {
	result = &v1alpha1.ScheduledScaler{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("scheduledscalers").
		Body(scheduledScaler).
		Do(context.TODO()).
		Into(result)
	return
}

// Update takes the representation of a scheduledScaler and updates it. Returns the server's representation of the scheduledScaler, and an error, if there is any.
func (c *scheduledScalers) Update(scheduledScaler *v1alpha1.ScheduledScaler) (result *v1alpha1.ScheduledScaler, err error) {
	result = &v1alpha1.ScheduledScaler{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("scheduledscalers").
		Name(scheduledScaler.Name).
		Body(scheduledScaler).
		Do(context.TODO()).
		Into(result)
	return
}

// Delete takes name of the scheduledScaler and deletes it. Returns an error if one occurs.
func (c *scheduledScalers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scheduledscalers").
		Name(name).
		Body(options).
		Do(context.TODO()).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *scheduledScalers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scheduledscalers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do(context.TODO()).
		Error()
}

// Patch applies the patch and returns the patched scheduledScaler.
func (c *scheduledScalers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ScheduledScaler, err error) {
	result = &v1alpha1.ScheduledScaler{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("scheduledscalers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do(context.TODO()).
		Into(result)
	return
}
