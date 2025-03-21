//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	config "github.com/gardener/gardener-extension-runtime-gvisor/pkg/apis/config"
	apisconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	componentbaseconfig "k8s.io/component-base/config"
	configv1alpha1 "k8s.io/component-base/config/v1alpha1"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*ControllerConfiguration)(nil), (*config.ControllerConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ControllerConfiguration_To_config_ControllerConfiguration(a.(*ControllerConfiguration), b.(*config.ControllerConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*config.ControllerConfiguration)(nil), (*ControllerConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_config_ControllerConfiguration_To_v1alpha1_ControllerConfiguration(a.(*config.ControllerConfiguration), b.(*ControllerConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*GVisorConfiguration)(nil), (*config.GVisorConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_GVisorConfiguration_To_config_GVisorConfiguration(a.(*GVisorConfiguration), b.(*config.GVisorConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*config.GVisorConfiguration)(nil), (*GVisorConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_config_GVisorConfiguration_To_v1alpha1_GVisorConfiguration(a.(*config.GVisorConfiguration), b.(*GVisorConfiguration), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_ControllerConfiguration_To_config_ControllerConfiguration(in *ControllerConfiguration, out *config.ControllerConfiguration, s conversion.Scope) error {
	if in.ClientConnection != nil {
		in, out := &in.ClientConnection, &out.ClientConnection
		*out = new(componentbaseconfig.ClientConnectionConfiguration)
		if err := configv1alpha1.Convert_v1alpha1_ClientConnectionConfiguration_To_config_ClientConnectionConfiguration(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.ClientConnection = nil
	}
	out.HealthCheckConfig = (*apisconfigv1alpha1.HealthCheckConfig)(unsafe.Pointer(in.HealthCheckConfig))
	return nil
}

// Convert_v1alpha1_ControllerConfiguration_To_config_ControllerConfiguration is an autogenerated conversion function.
func Convert_v1alpha1_ControllerConfiguration_To_config_ControllerConfiguration(in *ControllerConfiguration, out *config.ControllerConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_ControllerConfiguration_To_config_ControllerConfiguration(in, out, s)
}

func autoConvert_config_ControllerConfiguration_To_v1alpha1_ControllerConfiguration(in *config.ControllerConfiguration, out *ControllerConfiguration, s conversion.Scope) error {
	if in.ClientConnection != nil {
		in, out := &in.ClientConnection, &out.ClientConnection
		*out = new(configv1alpha1.ClientConnectionConfiguration)
		if err := configv1alpha1.Convert_config_ClientConnectionConfiguration_To_v1alpha1_ClientConnectionConfiguration(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.ClientConnection = nil
	}
	out.HealthCheckConfig = (*apisconfigv1alpha1.HealthCheckConfig)(unsafe.Pointer(in.HealthCheckConfig))
	return nil
}

// Convert_config_ControllerConfiguration_To_v1alpha1_ControllerConfiguration is an autogenerated conversion function.
func Convert_config_ControllerConfiguration_To_v1alpha1_ControllerConfiguration(in *config.ControllerConfiguration, out *ControllerConfiguration, s conversion.Scope) error {
	return autoConvert_config_ControllerConfiguration_To_v1alpha1_ControllerConfiguration(in, out, s)
}

func autoConvert_v1alpha1_GVisorConfiguration_To_config_GVisorConfiguration(in *GVisorConfiguration, out *config.GVisorConfiguration, s conversion.Scope) error {
	out.ConfigFlags = (*map[string]string)(unsafe.Pointer(in.ConfigFlags))
	return nil
}

// Convert_v1alpha1_GVisorConfiguration_To_config_GVisorConfiguration is an autogenerated conversion function.
func Convert_v1alpha1_GVisorConfiguration_To_config_GVisorConfiguration(in *GVisorConfiguration, out *config.GVisorConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_GVisorConfiguration_To_config_GVisorConfiguration(in, out, s)
}

func autoConvert_config_GVisorConfiguration_To_v1alpha1_GVisorConfiguration(in *config.GVisorConfiguration, out *GVisorConfiguration, s conversion.Scope) error {
	out.ConfigFlags = (*map[string]string)(unsafe.Pointer(in.ConfigFlags))
	return nil
}

// Convert_config_GVisorConfiguration_To_v1alpha1_GVisorConfiguration is an autogenerated conversion function.
func Convert_config_GVisorConfiguration_To_v1alpha1_GVisorConfiguration(in *config.GVisorConfiguration, out *GVisorConfiguration, s conversion.Scope) error {
	return autoConvert_config_GVisorConfiguration_To_v1alpha1_GVisorConfiguration(in, out, s)
}
