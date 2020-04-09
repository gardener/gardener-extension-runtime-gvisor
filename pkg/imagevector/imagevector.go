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

//go:generate packr2

package imagevector

import (
	"fmt"
	"strings"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/version"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/gobuffalo/packr/v2"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var imageVector imagevector.ImageVector

func init() {
	box := packr.New("charts", "../../charts")

	imagesYaml, err := box.FindString("images.yaml")
	runtime.Must(err)

	imageVector, err = imagevector.Read(strings.NewReader(imagesYaml))
	runtime.Must(err)
	// image vector for components deployed by the gVisor extension
	imageVector, err = imagevector.WithEnvOverride(imageVector)
	runtime.Must(err)

	image, err := imageVector.FindImage(gvisor.RuntimeGVisorInstallationImageName)
	runtime.Must(err)
	fmt.Printf("Image %q - using image name: %q \n", gvisor.RuntimeGVisorInstallationImageName, image.String())
	runtime.Must(err)
}

// ImageVector is the image vector that contains all the needed images.
func ImageVector() imagevector.ImageVector {
	return imageVector
}

// FindImage returns the container runtime GVisor image.
func FindImage(name string) string {
	image, err := imageVector.FindImage(name)
	runtime.Must(err)

	var (
		repository = image.String()
		tag        = version.Get().GitVersion
	)
	if image.Tag != nil {
		repository = image.Repository
		tag = *image.Tag
	}
	calculatedImage := imagevector.Image{
		Repository: repository,
		Tag:        &tag,
	}
	return calculatedImage.String()
}
