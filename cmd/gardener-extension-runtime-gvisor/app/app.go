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

package app

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/gardener/extensions/pkg/controller"
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	"github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	heartbeatcmd "github.com/gardener/gardener/extensions/pkg/controller/heartbeat/cmd"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/component-base/version/verflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	gvisorcontroller "github.com/gardener/gardener-extension-runtime-gvisor/pkg/controller"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/gvisor"
	"github.com/gardener/gardener-extension-runtime-gvisor/pkg/healthcheck"
)

// NewControllerManagerCommand creates a new command that is used to start the Container runtime gvisor controller.
func NewControllerManagerCommand(ctx context.Context) *cobra.Command {
	var (
		generalOpts = &controllercmd.GeneralOptions{}
		restOpts    = &controllercmd.RESTOptions{}
		mgrOpts     = &controllercmd.ManagerOptions{
			LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
			LeaderElection:             true,
			LeaderElectionID:           controllercmd.LeaderElectionNameID(gvisor.Name),
			LeaderElectionNamespace:    os.Getenv("LEADER_ELECTION_NAMESPACE"),
		}
		reconcileOpts = &controllercmd.ReconcilerOptions{
			IgnoreOperationAnnotation: true,
		}

		// options for the runtime-gvisor controller
		gvisorCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}

		// options for the health care controller
		healthCheckCtrlOpts = &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		}

		// options for the heartbeat controller
		heartbeatCtrlOpts = &heartbeatcmd.Options{
			ExtensionName:        gvisor.Name,
			RenewIntervalSeconds: 30,
			Namespace:            os.Getenv("LEADER_ELECTION_NAMESPACE"),
		}

		aggOption = controllercmd.NewOptionAggregator(
			generalOpts,
			restOpts,
			mgrOpts,
			gvisorCtrlOpts,
			controllercmd.PrefixOption("healthcheck-", healthCheckCtrlOpts),
			controllercmd.PrefixOption("heartbeat-", heartbeatCtrlOpts),
			reconcileOpts,
		)
	)

	cmd := &cobra.Command{
		Use: fmt.Sprintf("%s-controller-manager", gvisor.Name),

		RunE: func(cmd *cobra.Command, args []string) error {
			// Act on version flag, if one was specified
			verflag.PrintAndExitIfRequested()

			if err := aggOption.Complete(); err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}

			if err := heartbeatCtrlOpts.Validate(); err != nil {
				return err
			}

			completedMgrOpts := mgrOpts.Completed().Options()
			completedMgrOpts.ClientDisableCacheFor = []client.Object{
				&corev1.Secret{}, // applied for ManagedResources
			}

			mgr, err := manager.New(restOpts.Completed().Config, completedMgrOpts)
			if err != nil {
				return fmt.Errorf("could not instantiate manager: %w", err)
			}

			if err := controller.AddToScheme(mgr.GetScheme()); err != nil {
				return fmt.Errorf("could not update manager scheme: %w", err)
			}

			reconcileOpts.Completed().Apply(&gvisorcontroller.DefaultAddOptions.IgnoreOperationAnnotation)
			gvisorCtrlOpts.Completed().Apply(&gvisorcontroller.DefaultAddOptions.Controller)
			heartbeatCtrlOpts.Completed().Apply(&heartbeat.DefaultAddOptions)

			if err := gvisorcontroller.AddToManager(ctx, mgr); err != nil {
				return fmt.Errorf("could not add controllers to manager: %w", err)
			}

			if err := healthcheck.AddToManager(ctx, mgr); err != nil {
				return fmt.Errorf("could not add health check controller to manager: %w", err)
			}

			if err := heartbeat.AddToManager(ctx, mgr); err != nil {
				return fmt.Errorf("could not add heartbeat controller to manager: %w", err)
			}

			if err := mgr.Start(ctx); err != nil {
				return fmt.Errorf("error running manager: %w", err)
			}

			return nil
		},
	}

	verflag.AddFlags(cmd.Flags())
	aggOption.AddFlags(cmd.Flags())

	return cmd
}
