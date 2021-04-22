// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package operatingsystemconfig

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/controllerutils"
)

// reconciler reconciles OperatingSystemConfig resources of Gardener's `extensions.gardener.cloud`
// API group.
type reconciler struct {
	logger   logr.Logger
	actuator Actuator

	client        client.Client
	reader        client.Reader
	scheme        *runtime.Scheme
	statusUpdater extensionscontroller.StatusUpdater
}

// NewReconciler creates a new reconcile.Reconciler that reconciles
// OperatingSystemConfig resources of Gardener's `extensions.gardener.cloud` API group.
func NewReconciler(actuator Actuator) reconcile.Reconciler {
	logger := log.Log.WithName(ControllerName)

	return extensionscontroller.OperationAnnotationWrapper(
		func() client.Object { return &extensionsv1alpha1.OperatingSystemConfig{} },
		&reconciler{
			logger:        logger,
			actuator:      actuator,
			statusUpdater: extensionscontroller.NewStatusUpdater(logger),
		},
	)
}

// InjectFunc enables dependency injection into the actuator.
func (r *reconciler) InjectFunc(f inject.Func) error {
	return f(r.actuator)
}

// InjectClient injects the controller runtime client into the reconciler.
func (r *reconciler) InjectClient(client client.Client) error {
	r.client = client
	r.statusUpdater.InjectClient(client)
	return nil
}

func (r *reconciler) InjectAPIReader(reader client.Reader) error {
	r.reader = reader
	return nil
}

func (r *reconciler) InjectScheme(scheme *runtime.Scheme) error {
	r.scheme = scheme
	return nil
}

// Reconcile is the reconciler function that gets executed in case there are new events for the `OperatingSystemConfig`
// resources.
func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	osc := &extensionsv1alpha1.OperatingSystemConfig{}
	if err := r.client.Get(ctx, request.NamespacedName, osc); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("could not fetch OperatingSystemConfig: %+v", err)
	}

	shoot, err := extensionscontroller.GetShoot(ctx, r.client, request.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	if extensionscontroller.IsShootFailed(shoot) {
		r.logger.Info("Stop reconciling OperatingSystemConfig of failed Shoot.", "namespace", request.Namespace, "name", osc.Name)
		return reconcile.Result{}, nil
	}

	operationType := gardencorev1beta1helper.ComputeOperationType(osc.ObjectMeta, osc.Status.LastOperation)

	switch {
	case extensionscontroller.IsMigrated(osc):
		return reconcile.Result{}, nil
	case operationType == gardencorev1beta1.LastOperationTypeMigrate:
		return r.migrate(ctx, osc)
	case osc.DeletionTimestamp != nil:
		return r.delete(ctx, osc)
	case osc.Annotations[v1beta1constants.GardenerOperation] == v1beta1constants.GardenerOperationRestore:
		return r.restore(ctx, osc)
	default:
		return r.reconcile(ctx, osc, operationType)
	}
}

func (r *reconciler) reconcile(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig, operationType gardencorev1beta1.LastOperationType) (reconcile.Result, error) {
	if err := controllerutils.EnsureFinalizer(ctx, r.reader, r.client, osc, FinalizerName); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.statusUpdater.Processing(ctx, osc, operationType, "Reconciling the operating system config"); err != nil {
		return reconcile.Result{}, err
	}

	userData, command, units, err := r.actuator.Reconcile(ctx, osc)
	if err != nil {
		_ = r.statusUpdater.Error(ctx, osc, extensionscontroller.ReconcileErrCauseOrErr(err), operationType, "Error reconciling operating system config")
		return extensionscontroller.ReconcileErr(err)
	}

	secret, err := r.createOrUpdateOSCResultSecret(ctx, osc, userData)
	if err != nil {
		_ = r.statusUpdater.Error(ctx, osc, extensionscontroller.ReconcileErrCauseOrErr(err), operationType, "Could not apply secret for generated cloud config")
		return extensionscontroller.ReconcileErr(err)
	}

	setOSCStatus(osc, secret, command, units)

	if err := r.statusUpdater.Success(ctx, osc, operationType, "Successfully reconciled operating system config"); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) restore(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) (reconcile.Result, error) {
	if err := controllerutils.EnsureFinalizer(ctx, r.reader, r.client, osc, FinalizerName); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.statusUpdater.Processing(ctx, osc, gardencorev1beta1.LastOperationTypeRestore, "Restoring the operating system config"); err != nil {
		return reconcile.Result{}, err
	}

	userData, command, units, err := r.actuator.Restore(ctx, osc)
	if err != nil {
		_ = r.statusUpdater.Error(ctx, osc, extensionscontroller.ReconcileErrCauseOrErr(err), gardencorev1beta1.LastOperationTypeRestore, "Error restoring operating system config")
		return extensionscontroller.ReconcileErr(err)
	}

	secret, err := r.createOrUpdateOSCResultSecret(ctx, osc, userData)
	if err != nil {
		_ = r.statusUpdater.Error(ctx, osc, extensionscontroller.ReconcileErrCauseOrErr(err), gardencorev1beta1.LastOperationTypeRestore, "Could not apply secret for generated cloud config")
		return extensionscontroller.ReconcileErr(err)
	}

	setOSCStatus(osc, secret, command, units)

	if err := r.statusUpdater.Success(ctx, osc, gardencorev1beta1.LastOperationTypeRestore, "Successfully restored operating system config"); err != nil {
		return reconcile.Result{}, err
	}

	if err := extensionscontroller.RemoveAnnotation(ctx, r.client, osc, v1beta1constants.GardenerOperation); err != nil {
		return reconcile.Result{}, fmt.Errorf("error removing annotation from OperationSystemConfig: %+v", err)
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) delete(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(osc, FinalizerName) {
		r.logger.Info("Deleting operating system config causes a no-op as there is no finalizer.", "osc", osc.Name)
		return reconcile.Result{}, nil
	}

	if err := r.statusUpdater.Processing(ctx, osc, gardencorev1beta1.LastOperationTypeDelete, "Deleting the operating system config"); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.actuator.Delete(ctx, osc); err != nil {
		_ = r.statusUpdater.Error(ctx, osc, extensionscontroller.ReconcileErrCauseOrErr(err), gardencorev1beta1.LastOperationTypeDelete, "Error deleting operating system config")
		return extensionscontroller.ReconcileErr(err)
	}

	if err := r.statusUpdater.Success(ctx, osc, gardencorev1beta1.LastOperationTypeDelete, "Successfully deleted operating system config"); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Removing finalizer.", "osc", osc.Name)
	if err := controllerutils.RemoveFinalizer(ctx, r.reader, r.client, osc, FinalizerName); err != nil {
		return reconcile.Result{}, fmt.Errorf("error removing finalizer from OperationSystemConfig: %+v", err)
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) migrate(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(osc, FinalizerName) {
		r.logger.Info("Migrating operating system config causes a no-op as there is no finalizer.", "osc", osc.Name)
		return reconcile.Result{}, nil
	}

	if err := r.statusUpdater.Processing(ctx, osc, gardencorev1beta1.LastOperationTypeMigrate, "Migrating the operating system config"); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.actuator.Migrate(ctx, osc); err != nil {
		_ = r.statusUpdater.Error(ctx, osc, extensionscontroller.ReconcileErrCauseOrErr(err), gardencorev1beta1.LastOperationTypeMigrate, "Error migrating operating system config")
		return extensionscontroller.ReconcileErr(err)
	}

	if err := r.statusUpdater.Success(ctx, osc, gardencorev1beta1.LastOperationTypeMigrate, "Successfully migrated operating system config"); err != nil {
		return reconcile.Result{}, err
	}

	r.logger.Info("Removing finalizer.", "osc", osc.Name)
	if err := extensionscontroller.DeleteAllFinalizers(ctx, r.client, osc); err != nil {
		return reconcile.Result{}, fmt.Errorf("Error removing all finalizers from OperationSystemConfig: %+v", err)
	}

	if err := extensionscontroller.RemoveAnnotation(ctx, r.client, osc, v1beta1constants.GardenerOperation); err != nil {
		return reconcile.Result{}, fmt.Errorf("error removing annotation from OperationSystemConfig: %+v", err)
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) createOrUpdateOSCResultSecret(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig, userData []byte) (*corev1.Secret, error) {
	secret := &corev1.Secret{ObjectMeta: SecretObjectMetaForConfig(osc)}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.client, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[extensionsv1alpha1.OperatingSystemConfigSecretDataKey] = userData
		return controllerutil.SetControllerReference(osc, secret, r.scheme)
	}); err != nil {
		return nil, err
	}
	return secret, nil
}

func setOSCStatus(osc *extensionsv1alpha1.OperatingSystemConfig, secret *corev1.Secret, command *string, units []string) {
	osc.Status.CloudConfig = &extensionsv1alpha1.CloudConfig{
		SecretRef: corev1.SecretReference{
			Name:      secret.Name,
			Namespace: secret.Namespace,
		},
	}
	osc.Status.Units = units
	if command != nil {
		osc.Status.Command = command
	}
}
