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

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE k8s.restdev.com/operators/pkg/client/clientset/versioned/typed/scaling/v1alpha1 ScalingV1alpha1Interface

import (
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
	v1alpha1 "k8s.restdev.com/operators/pkg/apis/scaling/v1alpha1"
)

type ScalingV1alpha1Interface interface {
	RESTClient() rest.Interface
	ScheduledScalersGetter
}

// ScalingV1alpha1Client is used to interact with features provided by the scaling.k8s.restdev.com group.
type ScalingV1alpha1Client struct {
	restClient rest.Interface
}

func (c *ScalingV1alpha1Client) ScheduledScalers(namespace string) ScheduledScalerInterface {
	return newScheduledScalers(c, namespace)
}

// NewForConfig creates a new ScalingV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*ScalingV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ScalingV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new ScalingV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *ScalingV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new ScalingV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *ScalingV1alpha1Client {
	return &ScalingV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.CodecFactory{}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ScalingV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
