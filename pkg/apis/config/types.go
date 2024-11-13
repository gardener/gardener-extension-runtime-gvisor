package config

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GVisorConfiguration defines the configuration for the GVisor runtime extension.
type GVisorConfiguration struct {
	metav1.TypeMeta

	// Capabilities to the granted to GVisor containers.
	AdditionalCapabilities *GVisorAdditionalCapabilities
}

// GVisorAdditionalCapabilities is the list of capabilities that can be granted to GVisor containers.
type GVisorAdditionalCapabilities struct {
	// Allows the process to bind to any address within the available namespaces
	CapabilityNetRaw *bool
	// Grants all Capabilities to the process
	CapabilitySysAdmin *bool
}
