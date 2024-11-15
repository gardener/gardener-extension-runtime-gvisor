package config

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GVisorConfiguration defines the configuration for the GVisor runtime extension.
type GVisorConfiguration struct {
	metav1.TypeMeta

	// ConfigFlags is a map of additional flags that are passed to the runsc binary.
	ConfigFlags *map[string]string
}
