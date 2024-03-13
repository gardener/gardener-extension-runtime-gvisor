// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"

	extensioncontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/containerruntime"
	"github.com/gardener/gardener/extensions/pkg/util"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the Gvisor container runtime controller to the manager.
type AddOptions struct {
	// Controller are the controller.Options.
	Controller controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
func AddToManagerWithOptions(ctx context.Context, mgr manager.Manager, opts AddOptions) error {
	scheme := mgr.GetScheme()
	if err := resourcesv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}

	return containerruntime.Add(ctx, mgr, containerruntime.AddArgs{
		Actuator:                  NewActuator(mgr.GetClient(), extensioncontroller.ChartRendererFactoryFunc(util.NewChartRendererForShoot)),
		ControllerOptions:         opts.Controller,
		Predicates:                containerruntime.DefaultPredicates(ctx, mgr, opts.IgnoreOperationAnnotation),
		Type:                      gvisor.Type,
		IgnoreOperationAnnotation: opts.IgnoreOperationAnnotation,
	})
}

// AddToManager adds a controller with the default Options.
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return AddToManagerWithOptions(ctx, mgr, DefaultAddOptions)
}
