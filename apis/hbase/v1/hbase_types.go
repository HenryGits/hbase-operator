/*
Copyright 2021.

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

package v1

import (
	"gitee.com/dmcca/gotools/compass/typed"
	typedv1 "gitee.com/dmcca/gotools/compass/typed/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+genclient
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Hbase is the Schema for the hbases API
type Hbase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HbaseSpec   `json:"spec,omitempty"`
	Status HbaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HbaseList contains a list of Hbase
type HbaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hbase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hbase{}, &HbaseList{})
}

// HbaseStatus defines the observed state of Hbase
type HbaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// HbaseSpec defines the desired state of Hbase
type HbaseSpec struct {
	// MasterSpec is definition of HBase Master server
	// +optional
	MasterSpec ServerSpec `json:"masterSpec,omitempty"`
	// RegionServerSpec is definition of HBase RegionServer
	// +optional
	RegionServerSpec ServerSpec `json:"regionServerSpec,omitempty"`
	// +optional
	ThriftServer ServerSpec `json:"thriftServer,omitempty"`
	// +optional
	Image typedv1.Image `json:"image,omitempty"`
}

// ServerSpec is a specification for an HBase server (Master or Regionserver)
type ServerSpec struct {
	// +optional
	Volume typed.Volume `json:"volume,omitempty"`
	// 实例个数
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// +optional
	Port typedv1.Port `json:"port,omitempty"`
}
