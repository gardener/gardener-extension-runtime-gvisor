package config

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GvisorConfiguration struct {
	metav1.TypeMeta

	// Capabilities to the granted to GVisor containers.
	AdditionalCapabilities *GVisorAdditionalCapabilities
}

// List of capabilities that can be granted to GVisor containers.
type GVisorAdditionalCapabilities struct {
	CapabilityNetRaw   *bool
	CapabilitySysAdmin *bool
}
