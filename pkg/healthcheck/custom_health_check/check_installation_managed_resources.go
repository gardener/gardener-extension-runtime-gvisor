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

package general

import (
	"context"
	"fmt"

	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck/general"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/controller"
)

// GVisorInstallationManagedResourcesHealthChecker contains all the information for the ManagedResource HealthCheck
type GVisorInstallationManagedResourcesHealthChecker struct {
	logger      logr.Logger
	seedClient  client.Client
	shootClient client.Client
}

// CheckGVisorInstallationManagedResources is a healthCheck function to check ManagedResources
func CheckGVisorInstallationManagedResources() healthcheck.HealthCheck {
	return &GVisorInstallationManagedResourcesHealthChecker{}
}

// InjectSeedClient injects the seed client
func (healthChecker *GVisorInstallationManagedResourcesHealthChecker) InjectSeedClient(seedClient client.Client) {
	healthChecker.seedClient = seedClient
}

// InjectShootClient injects the shoot client
func (healthChecker *GVisorInstallationManagedResourcesHealthChecker) InjectShootClient(shootClient client.Client) {
	healthChecker.shootClient = shootClient
}

// SetLoggerSuffix injects the logger
func (healthChecker *GVisorInstallationManagedResourcesHealthChecker) SetLoggerSuffix(provider, extension string) {
	healthChecker.logger = log.Log.WithName(fmt.Sprintf("%s-%s-healthcheck-managed-resource", provider, extension))
}

// DeepCopy clones the healthCheck struct by making a copy and returning the pointer to that new copy
func (healthChecker *GVisorInstallationManagedResourcesHealthChecker) DeepCopy() healthcheck.HealthCheck {
	copy := *healthChecker
	return &copy
}

// Check executes the health check
func (healthChecker *GVisorInstallationManagedResourcesHealthChecker) Check(ctx context.Context, request types.NamespacedName) (*healthcheck.SingleCheckResult, error) {
	cr := &extensionsv1alpha1.ContainerRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.Name,
			Namespace: request.Namespace,
		},
	}
	if err := healthChecker.seedClient.Get(ctx, client.ObjectKeyFromObject(cr), cr); err != nil {
		return nil, err
	}

	// compute the managed resource name
	managedResourceInstallationName := controller.GetGVisorInstallationManagedResourceName(cr)

	checker := general.CheckManagedResource(managedResourceInstallationName)
	mrChecker, ok := checker.(*general.ManagedResourceHealthChecker)
	if !ok {
		return nil, fmt.Errorf("cannot construct managed resource health checker")
	}
	mrChecker.InjectSeedClient(healthChecker.seedClient)
	mrChecker.SetLoggerSuffix("", extensionsv1alpha1.ContainerRuntimeResource)

	return mrChecker.Check(ctx, types.NamespacedName{
		Namespace: request.Namespace,
	})
}
