// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package imagevector

import (
	_ "embed"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/version"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
)

var (
	//go:embed images.yaml
	ImagesYAML  string
	imageVector imagevector.ImageVector
)

func init() {
	var err error

	imageVector, err = imagevector.Read([]byte(ImagesYAML))
	runtime.Must(err)
	// image vector for components deployed by the gVisor extension
	imageVector, err = imagevector.WithEnvOverride(imageVector)
	runtime.Must(err)

	_, err = imageVector.FindImage(gvisor.RuntimeGVisorInstallationImageName)
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
