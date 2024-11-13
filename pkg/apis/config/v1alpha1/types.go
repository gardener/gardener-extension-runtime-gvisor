package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GVisorConfiguration defines the configuration for the GVisor runtime extension.
type GVisorConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// Capabilities to the granted to GVisor containers.
	// +optional
	AdditionalCapabilities *GVisorAdditionalCapabilities `json:"additionalCapabilities,omitempty"`
}

// GVisorAdditionalCapabilities is the list of capabilities that can be granted to GVisor containers.
type GVisorAdditionalCapabilities struct {
	// Allows the process to bind to any address within the available namespaces
	// +optional
	CapabilityNetRaw *bool `json:"NET_RAW,omitempty"`
	// Grants all Capabilities to the process
	// +optional
	CapabilitySysAdmin *bool `json:"SYS_ADMIN,omitempty"`
}
