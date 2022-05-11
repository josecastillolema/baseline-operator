/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	perfv1 "github.com/josecastillolema/baseline-operator/api/v1"
)

// BaselineReconciler reconciles a Baseline object
type BaselineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=perf.baseline.io,resources=baselines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=perf.baseline.io,resources=baselines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=perf.baseline.io,resources=baselines/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *BaselineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the Baseline instance
	baseline := &perfv1.Baseline{}
	err := r.Get(ctx, req.NamespacedName, baseline)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Baseline resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Baseline")
		return ctrl.Result{}, err
	}

	// Check if the daemonset already exists, if not create a new one
	found := &appsv1.DaemonSet{}
	err = r.Get(ctx, types.NamespacedName{Name: baseline.Name, Namespace: baseline.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new daemonset
		dep := r.daemonsetForBaseline(baseline)
		log.Info("Creating a new DaemonSet", "DaemonSet.Namespace", dep.Namespace, "DaemonSet.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new DaemonSet", "DaemonSet.Namespace", dep.Namespace, "DaemonSet.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get DaemonSet")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// daemonsetForBaseline returns a baseline DaemonSet object
func (r *BaselineReconciler) daemonsetForBaseline(m *perfv1.Baseline) *appsv1.DaemonSet {
	ls := labelsForBaseline(m.Name)

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "memcached:1.4.36-alpine",
						Name:    "memcached",
						Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "memcached",
						}},
					}},
				},
			},
		},
	}
	// Set Baseline instance as the owner and controller
	ctrl.SetControllerReference(m, ds, r.Scheme)
	return ds
}

// labelsForBaseline returns the labels for selecting the resources
// belonging to the given baseline CR name.
func labelsForBaseline(name string) map[string]string {
	return map[string]string{"app": "baseline", "baseline_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *BaselineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&perfv1.Baseline{}).
		Owns(&appsv1.DaemonSet{}).
		Complete(r)
}
