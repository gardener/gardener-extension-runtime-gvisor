// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GVisorConfiguration defines the configuration for the gVisor runtime resource.
type GVisorConfiguration struct {
	metav1.TypeMeta

	// ConfigFlags is a map of additional flags that are passed to the runsc binary.
	ConfigFlags *map[string]string

	// TestImageTag is the tag for the gardener-extension-runtime-gvisor-installation image to be tested.
	// It requires that the `gvisorInstallation.testRepository` is configured in the operator extension values and
	// the image has been uploaded and tagged accordingly.
	// Only used for development and testing purposes. Does not work if `gvisorInstallation.testRepository` is not specified.
	TestImageTag *string
}
