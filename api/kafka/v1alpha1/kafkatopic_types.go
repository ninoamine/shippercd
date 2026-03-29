/*
Copyright 2026.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KafkaTopicSpec defines the desired state of KafkaTopic
type KafkaTopicSpec struct {
	// Name is the topic name in Kafka
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Partitions is the number of partitions for the topic
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	Partitions int `json:"partitions"`
	// ReplicationFactor is the replication factor for the topic
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	ReplicationFactor int `json:"replicationFactor"`
}

// KafkaTopicStatus defines the observed state of KafkaTopic.
type KafkaTopicStatus struct {
	// Conditions represent the latest available observations of the topic's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// +optional
	Ready bool `json:"ready"`
	// +optional
	Message string `json:"message"`
	// +optional
	Error string `json:"error"`
	// +optional
	CreatedResources []string `json:"createdResources"`
	// +optional
	DeletedResources []string `json:"deletedResources"`
	// +optional
	UpdatedResources []string `json:"updatedResources"`
	// +optional
	FailedResources []string `json:"failedResources"`
	// +optional
	PendingResources []string `json:"pendingResources"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KafkaTopic is the Schema for the kafkatopics API
type KafkaTopic struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of KafkaTopic
	// +required
	Spec KafkaTopicSpec `json:"spec"`

	// status defines the observed state of KafkaTopic
	// +optional
	Status KafkaTopicStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// KafkaTopicList contains a list of KafkaTopic
type KafkaTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []KafkaTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaTopic{}, &KafkaTopicList{})
}
