// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gvisor

import "path/filepath"

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
	// ChartsPath is the path to the charts
	ChartsPath = filepath.Join("charts")
	// InternalChartsPath is the path to the internal charts
	InternalChartsPath = filepath.Join(ChartsPath, "internal")
	// InstallationChartPath path for internal GVisor installation Chart
	InstallationChartPath = filepath.Join(InternalChartsPath, "gvisor-installation")
	// InstallationChartPath path for internal GVisor Chart
	ChartPath = filepath.Join(InternalChartsPath, "gvisor")
)
