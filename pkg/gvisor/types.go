// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package gvisor

import (
	"path/filepath"

	"github.com/gardener/gardener-extension-runtime-gvisor/charts"
)

const (
	// Name is a constant to identify the gVisor extension
	Name = "runtime-gvisor"

	// RuntimeGVisorInstallationImageName is the image name for gvisor installation chart
	RuntimeGVisorInstallationImageName = "runtime-gvisor-installation"

	// InstallationReleaseName is the name of the gVisor installation chart
	InstallationReleaseName = "gvisor-installation"
	// ReleaseName is the name of the gVisor chart
	ReleaseName = "gvisor"
)

var (
	// InstallationChartPath path for internal GVisor installation Chart
	InstallationChartPath = filepath.Join(charts.InternalChartsPath, "gvisor-installation")
	// ChartPath is the path for internal GVisor Chart.
	ChartPath = filepath.Join(charts.InternalChartsPath, "gvisor")
)
