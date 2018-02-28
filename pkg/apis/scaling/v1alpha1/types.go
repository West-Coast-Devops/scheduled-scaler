/*
Copyright 2017 The Kubernetes Authors.
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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScheduledScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScheduledScalerSpec   `json:"spec"`
	Status ScheduledScalerStatus `json:"status"`
}

type ScheduledScalerStep struct {
	Runat string `json:"runat"`
	Mode string `json:"mode"`
	MinReplicas *int32 `json:"minReplicas"`
	MaxReplicas *int32 `json:"maxReplicas"`
	Replicas *int32 `json:"replicas"`
}

type ScheduledScalerSpec struct {
	Target struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"target"`
	Steps []ScheduledScalerStep `json:"steps"`
	TimeZone string `json:"timeZone"`
}

// FooStatus is the status for a Foo resource
type ScheduledScalerStatus struct {
	Mode string `json:"mode"`
	MinReplicas int32 `json:"minReplicas"`
	MaxReplicas int32 `json:"maxReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ScheduledScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ScheduledScaler `json:"items"`
}
