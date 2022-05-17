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
	"fmt"
	"reflect"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

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
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=perf.baseline.io,resources=baselines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=perf.baseline.io,resources=baselines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=perf.baseline.io,resources=baselines/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=perf.baseline.io,resources=events,verbs=create;patch

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
		ds := r.daemonsetForBaseline(baseline)
		log.Info("Creating a new DaemonSet", "DaemonSet.Namespace", ds.Namespace, "DaemonSet.Name", ds.Name)
		err = r.Create(ctx, ds)
		if err != nil {
			log.Error(err, "Failed to create new DaemonSet", "DaemonSet.Namespace", ds.Namespace, "DaemonSet.Name", ds.Name)
			return ctrl.Result{}, err
		}
		// Daemonset created successfully - update status, return and requeue
		r.recorder.Event(baseline, "Normal", "Created", fmt.Sprintf("Created daemonset %s/%s", ds.Namespace, ds.Name))
		baseline.Status.Command = strings.Join(ds.Spec.Template.Spec.Containers[0].Command, " ")
		err := r.Status().Update(ctx, baseline)
		if err != nil {
			log.Error(err, "Failed to update Baseline status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get DaemonSet")
		return ctrl.Result{}, err
	}

	// Ensure the nodeSelector and tolerations are the same as the spec
	nodeSelector := baseline.Spec.NodeSelector
	tolerations := baseline.Spec.Tolerations
	image := baseline.Spec.Image
	if !reflect.DeepEqual(found.Spec.Template.Spec.NodeSelector, nodeSelector) ||
		!reflect.DeepEqual(found.Spec.Template.Spec.Tolerations, tolerations) ||
		found.Spec.Template.Spec.Containers[0].Image != image {
		found.Spec.Template.Spec.NodeSelector = nodeSelector
		found.Spec.Template.Spec.Tolerations = tolerations
		found.Spec.Template.Spec.Containers[0].Image = image
		log.Info("Updating the DaemonSet with the new spec", "DaemonSet.Namespace", found.Namespace, "DaemonSet.Name", found.Name)
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update DaemonSet", "DaemonSet.Namespace", found.Namespace, "DaemonSet.Name", found.Name)
			return ctrl.Result{}, err
		}
		r.recorder.Event(baseline, "Normal", "Updated", fmt.Sprintf("Updated daemonset %s/%s", found.Namespace, found.Name))
	}

	// Ensure the stressng parameters are the same as in the spec
	cpu := baseline.Spec.Cpu
	mem := baseline.Spec.Memory
	custom := baseline.Spec.Custom
	command := *&found.Spec.Template.Spec.Containers[0].Command
	updateCpu := ((cpu != 0) && !present(command, "--cpu", strconv.Itoa(int(cpu)), 1)) || ((cpu == 0) && strings.Contains(strings.Join(command, " "), "--cpu"))
	notMem := ((mem != "") && !present(command, "--vm", mem, 3)) || ((mem == "") && strings.Contains(strings.Join(command, " "), "--vm"))
	if updateCpu || notMem || !strings.Contains(strings.Join(command, " "), custom) {
		// Define a new daemonset
		ds := r.daemonsetForBaseline(baseline)
		log.Info("Recreating the DaemonSet with the new command", "DaemonSet.Namespace", ds.Namespace, "DaemonSet.Name", ds.Name)
		err = r.Delete(ctx, ds)
		if err != nil {
			log.Error(err, "Failed to delete previous DaemonSet", "DaemonSet.Namespace", ds.Namespace, "DaemonSet.Name", ds.Name)
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, ds)
		if err != nil {
			log.Error(err, "Failed to recreate DaemonSet", "DaemonSet.Namespace", ds.Namespace, "DaemonSet.Name", ds.Name)
			return ctrl.Result{}, err
		}
		// Daemonset recreated successfully - update status, return and requeue
		r.recorder.Event(baseline, "Normal", "Recreated", fmt.Sprintf("Rereated daemonset %s/%s", ds.Namespace, ds.Name))
		baseline.Status.Command = strings.Join(ds.Spec.Template.Spec.Containers[0].Command, " ")
		err := r.Status().Update(ctx, baseline)
		if err != nil {
			log.Error(err, "Failed to update Baseline status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// present returns if the item has the corresponding value
func present(commands []string, item string, value string, shift int) bool {
	if shift > len(commands) {
		return false
	}
	for i, n := range commands {
		if item == n {
			if (i+shift < len(commands)) && (commands[i+shift] == value) {
				return true
			}
		}
	}
	return false
}

// daemonsetForBaseline returns a baseline DaemonSet object
func (r *BaselineReconciler) daemonsetForBaseline(b *perfv1.Baseline) *appsv1.DaemonSet {
	ls := labelsForBaseline(b.Name)
	cpu := strconv.Itoa(int(b.Spec.Cpu))
	mem := b.Spec.Memory
	custom := b.Spec.Custom
	command := []string{"stress-ng", "--timeout", "0"}
	if cpu != "0" {
		command = append(command, "--cpu", cpu)
	}
	if mem != "" {
		command = append(command, "--vm", "1", "--vm-bytes", mem)
	}
	if custom != "" {
		command = append(command, strings.Split(custom, " ")...)
	}

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name,
			Namespace: b.Namespace,
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
					NodeSelector: b.Spec.NodeSelector,
					Tolerations:  b.Spec.Tolerations,
					Containers: []corev1.Container{{
						Image:   b.Spec.Image,
						Name:    "stressng",
						Command: command,
					}},
				},
			},
		},
	}
	// Set Baseline instance as the owner and controller
	ctrl.SetControllerReference(b, ds, r.Scheme)
	return ds
}

// labelsForBaseline returns the labels for selecting the resources
// belonging to the given baseline CR name.
func labelsForBaseline(name string) map[string]string {
	return map[string]string{"app": "baseline", "baseline_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *BaselineReconciler) SetupWithManager(mgr ctrl.Manager) error {

	r.recorder = mgr.GetEventRecorderFor("Baseline")

	return ctrl.NewControllerManagedBy(mgr).
		For(&perfv1.Baseline{}).
		Owns(&appsv1.DaemonSet{}).
		Complete(r)
}
