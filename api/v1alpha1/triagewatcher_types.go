/*
Copyright 2026 Akshay Raina.

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
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TriageWatcherSpec defines the desired state of TriageWatcher
// TriageWatcherSpec defines the desired state of a TriageWatcher monitoring entity.
type TriageWatcherSpec struct {
	// TargetNamespace dictates which specific namespace this worker instance monitors.
	// +kubebuilder:validation:Required
	TargetNamespace string `json:"targetNamespace"`

	// CooldownPeriod defines how long an ongoing anomaly is ignored after an alert ships. Example: "30m"
	// +kubebuilder:default:="30m"
	// +optional
	CooldownPeriod string `json:"cooldownPeriod,omitempty"`

	// SlackSecretRef provides the name of the Kubernetes Secret string containing your webhook endpoint.
	// +kubebuilder:validation:Required
	SlackSecretRef string `json:"slackSecretRef"`
}

// TriageWatcherStatus defines the observed state of an active TriageWatcher daemon loop.
type TriageWatcherStatus struct {
	// ActiveMonitors represents how many distinct workloads are currently under surveillance.
	ActiveMonitors int `json:"activeMonitors,omitempty"`

	// LastScrapeTime documents the timestamp of the last complete evaluation sequence sweep.
	LastScrapeTime string `json:"lastScrapeTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TriageWatcher is the Schema for the triagewatchers API
type TriageWatcher struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of TriageWatcher
	// +required
	Spec TriageWatcherSpec `json:"spec"`

	// status defines the observed state of TriageWatcher
	// +optional
	Status TriageWatcherStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// TriageWatcherList contains a list of TriageWatcher
type TriageWatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []TriageWatcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &TriageWatcher{}, &TriageWatcherList{})
		return nil
	})
}
