/*
Copyright 2025.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	PodGroupNameLabel  = "podgroup.jhwagner.github.io/name"
	PodGroupReadyLabel = "podgroup.jhwagner.github.io/ready"
)

// PodGroupReconciler watches Pods and manages pod groups
type PodGroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get

func (r *PodGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	groupName, hasPodGroupLabel := pod.Labels[PodGroupNameLabel]
	if !hasPodGroupLabel {
		return ctrl.Result{}, nil
	}

	log.Info("Processing pod in group", "podGroup", groupName)

	var podList corev1.PodList
	if err := r.List(ctx, &podList,
		client.InNamespace(req.Namespace),
		client.MatchingLabels{PodGroupNameLabel: groupName},
	); err != nil {
		log.Error(err, "Failed to list pods in group", "podGroup", groupName)
		return ctrl.Result{}, err
	}

	allRunning := true
	for _, p := range podList.Items {
		if p.Status.Phase != corev1.PodRunning {
			allRunning = false
			break
		}
	}

	if !allRunning {
		log.Info("Not all pods running yet", "podGroup", groupName, "totalPods", len(podList.Items))
		return ctrl.Result{}, nil
	}

	log.Info("All pods running, applying ready label", "podGroup", groupName, "totalPods", len(podList.Items))

	for _, p := range podList.Items {
		if p.Labels[PodGroupReadyLabel] == "true" {
			continue
		}

		podCopy := p.DeepCopy()
		if podCopy.Labels == nil {
			podCopy.Labels = make(map[string]string)
		}
		podCopy.Labels[PodGroupReadyLabel] = "true"

		if err := r.Update(ctx, podCopy); err != nil {
			log.Error(err, "Failed to update pod", "pod", p.Name)
			return ctrl.Result{}, err
		}
		log.Info("Applied ready label", "pod", p.Name, "podGroup", groupName)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(object client.Object) bool {
			pod := object.(*corev1.Pod)
			_, hasPodGroupLabel := pod.Labels[PodGroupNameLabel]
			return hasPodGroupLabel
		})).
		Complete(r)
}
