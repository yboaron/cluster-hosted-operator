/*
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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConfigSpec defines the desired state of Config
type ConfigSpec struct {
	LoadBalancer HaLoadBalanceConfig `json:"loadbalancer,omitempty"`
	DNS          DnsConfig           `json:"dns,omitempty"`
}

type HaLoadBalanceConfig struct {
	DefaultIngressHA EnableDisable `json:"defaultingressha,omitempty"`
	ApiLoadbalance   EnableDisable `json:"apiloadbalance,omitempty"`
}

type DnsConfig struct {
	NodesResolution EnableDisable `json:"nodesresolution,omitempty"`
	ApiResolution   EnableDisable `json:"apiresolution,omitempty"`
	AppsResolution  EnableDisable `json:"appsresolution,omitempty"`
}

// +kubebuilder:validation:Enum=Enable;Disable
type EnableDisable string

// ConfigStatus defines the observed state of Config
type ConfigStatus struct {
	IngressVipOwner string `json:"ingressvipowner,omitempty"`
	APIVipOwner     string `json:"apivipowner,omitempty"`

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Config is the Schema for the configs API
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}
