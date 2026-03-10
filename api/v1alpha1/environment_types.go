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

type PostgreSQLDatabase struct {
	Name string `json:"name"`
}

type MongoDBDatabase struct {
	Name string `json:"name"`
}

type KafkaTopic struct {
	Name string `json:"name"`
	Partitions int `json:"partitions"`
	ReplicationFactor int `json:"replicationFactor"`
}

// EnvironmentSpec defines the desired state of Environment
type EnvironmentSpec struct {
	Name string `json:"name"`
	// +optional
	PostgreSQLDatabases []PostgreSQLDatabase `json:"postgresqlDatabases"`
	// +optional
	MongoDBDatabases []MongoDBDatabase `json:"mongodbDatabases"`
	// +optional
	KafkaTopics []KafkaTopic `json:"kafkaTopics"`
}

// EnvironmentStatus defines the observed state of Environment.
type EnvironmentStatus struct {
	Ready bool `json:"ready"`
	Message string `json:"message"`
	Error string `json:"error"`
	CreatedResources []string `json:"createdResources"`
	DeletedResources []string `json:"deletedResources"`
	UpdatedResources []string `json:"updatedResources"`
	FailedResources []string `json:"failedResources"`
	PendingResources []string `json:"pendingResources"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Environment is the Schema for the environments API
type Environment struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Environment
	// +kubebuilder:validation:Required
	Spec EnvironmentSpec `json:"spec"`

	// status defines the observed state of Environment
	// +optional
	Status EnvironmentStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// EnvironmentList contains a list of Environment
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Environment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
