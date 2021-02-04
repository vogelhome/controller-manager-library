/*
 * SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"fmt"
	"strings"
	"time"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile"
	"github.com/gardener/controller-manager-library/pkg/controllermanager/controller/reconcile/reconcilers"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/resources"

	corev1 "k8s.io/api/core/v1"
)

var secretGK = resources.NewGroupKind("core", "Secret")

func init() {
	controller.Configure("cm").
		Reconciler(Create).RequireLease().
		DefaultWorkerPool(10, 0*time.Second).
		Commands("poll").
		StringOption("test", "Controller argument").
		OptionSource("test", controller.OptionSourceCreator(&Config{})).
		MainResource("core", "ConfigMap", controller.NamespaceSelection("default")).
		With(reconcilers.SecretUsageReconciler(controller.CLUSTER_MAIN)).
		MustRegister()
}

type Config struct {
	option string
}

func (this *Config) AddOptionsToSet(set config.OptionSet) {
	set.AddStringOption(&this.option, "option", "", "", "2nd controller argument")
}

func (this *Config) Prepare() error {
	if this.option == "abort" {
		return fmt.Errorf("test validation failed")
	}
	return nil
}

type reconciler struct {
	reconcile.DefaultReconciler
	controller controller.Interface
	secrets    *reconcilers.SecretUsageCache
}

var _ reconcile.Interface = &reconciler{}

///////////////////////////////////////////////////////////////////////////////

func Create(controller controller.Interface) (reconcile.Interface, error) {

	val, err := controller.GetStringOption("test")
	if err == nil {
		controller.Infof("found option test: %s", val)
	}

	config, err := controller.GetOptionSource("test")
	if err == nil {
		controller.Infof("found option option: %s", config.(*Config).option)
	}

	return &reconciler{
		controller: controller,
		secrets:    reconcilers.AccessSecretUsageCache(controller),
	}, nil
}

///////////////////////////////////////////////////////////////////////////////

func (h *reconciler) Setup() error {
	h.controller.Infof("setup of reconciler")
	return nil
}

func (h *reconciler) Start() {
	h.controller.EnqueueCommand("poll")
}

func (h *reconciler) Command(logger logger.LogContext, cmd string) reconcile.Status {
	logger.Infof("got command %q", cmd)
	return reconcile.Succeeded(logger).RescheduleAfter(10 * time.Second)
}

func (h *reconciler) Reconcile(logger logger.LogContext, obj resources.Object) reconcile.Status {
	switch o := obj.Data().(type) {
	case *corev1.ConfigMap:
		return h.reconcileConfigMap(logger, obj.ClusterKey(), o)
	}

	return reconcile.Succeeded(logger)
}

func (h *reconciler) Delete(logger logger.LogContext, obj resources.Object) reconcile.Status {
	// logger.Infof("delete infrastructure %s", resources.Description(obj))
	logger.Infof("should delete")
	return reconcile.Succeeded(logger)
}

func (h *reconciler) Deleted(logger logger.LogContext, key resources.ClusterObjectKey) reconcile.Status {
	// logger.Infof("delete infrastructure %s", resources.Description(obj))
	logger.Infof("is deleted")
	h.secrets.SetUsesFor(key, nil)
	return reconcile.Succeeded(logger)
}

func (h *reconciler) secretRef(namespace, name string) *resources.ClusterObjectKey {
	if strings.HasPrefix(name, "secret") {
		name = name[6:]
		if name != "" {
			key := resources.NewClusterKey(h.controller.GetMainCluster().GetId(), secretGK, namespace, name)
			return &key
		}
	}
	return nil
}

func (h *reconciler) reconcileConfigMap(logger logger.LogContext, key resources.ClusterObjectKey, configMap *corev1.ConfigMap) reconcile.Status {
	logger.Infof("should reconcile configmap")
	// Example how to add to workqueue
	// resources, _ := h.controller.Data(configMap)
	// key, _ := controller.ObjectKeyFunc(resources)
	//	h.controller.GetWorkqueue().Add(key)

	secret := h.secretRef(configMap.Namespace, configMap.Name)
	h.secrets.SetUsesFor(key, secret)
	return reconcile.Succeeded(logger)
}
