package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GvisorConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// Capabilities to the granted to GVisor containers.
	// +optional
	AdditionalCapabilities *GVisorAdditionalCapabilities `json:"additionalCapabilities,omitempty"`
}

// List of capabilities that can be granted to GVisor containers.
type GVisorAdditionalCapabilities struct {
	// +optional
	CapabilityNetRaw *bool `json:"NET_RAW,omitempty"`
	// +optional
	CapabilitySysAdmin *bool `json:"SYS_ADMIN,omitempty"`
}
