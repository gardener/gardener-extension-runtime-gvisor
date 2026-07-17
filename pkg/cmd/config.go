// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/pflag"
)

// ConfigOptions are command line options that can be set for the Config.
type ConfigOptions struct {
	// InstallationTestRepository is the repository for test images of the gardener-extension-runtime-gvisor-installation container
	InstallationTestRepository string

	config *Config
}

// Config is a completed controller configuration.
type Config struct {
	// InstallationTestRepository is the repository for test images of the gardener-extension-runtime-gvisor-installation container
	InstallationTestRepository *string
}

// Complete implements Completer.Complete.
func (c *ConfigOptions) Complete() error {
	c.config = &Config{}
	if c.InstallationTestRepository != "" {
		c.config.InstallationTestRepository = &c.InstallationTestRepository
	}
	return nil
}

// Completed returns the completed Config. Only call this if `Complete` was successful.
func (c *ConfigOptions) Completed() *Config {
	return c.config
}

// AddFlags implements Flagger.AddFlags.
func (c *ConfigOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.InstallationTestRepository, "gvisor-installation-test-repository", "", "repository with test images of the gardener-extension-runtime-gvisor-installation container")
}

// Apply sets the values of this Config in the given Config.
func (c *Config) Apply(destCfg *Config) {
	*destCfg = *c
}
