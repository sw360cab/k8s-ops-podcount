/*
Copyright 2023.

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
	"sigs.k8s.io/controller-runtime/pkg/log"

	podcountv1 "github.com/k8s-ops-hello/api/v1"
)

// PodCountReconciler reconciles a PodCount object
type PodCountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pod-count.minimalgap.com,resources=podcounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pod-count.minimalgap.com,resources=podcounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pod-count.minimalgap.com,resources=podcounts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PodCount object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *PodCountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("reconciling foo custom resource")

	// Get the Foo resource that triggered the reconciliation request
	var podcount podcountv1.PodCount
	if err := r.Get(ctx, req.NamespacedName, &podcount); err != nil {
		log.Error(err, "unable to fetch CRD of counter")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get all the nodes
	var nodes corev1.NodeList
	var pods corev1.PodList
	if err := r.List(ctx, &nodes); err != nil {
		log.Error(err, "unable to list nodes")
	} else {
		for _, node := range nodes.Items {
			opts := []client.ListOption{
				// client.InNamespace(request.NamespacedName.Namespace),
				client.MatchingLabels{"name": "ok"},
				client.MatchingFields{"spec.nodeName": node.Name},
			}

			// get pods in node
			if err := r.List(ctx, &pods, opts...); err != nil {
				log.Error(err, "unable to list pods")
			} else {
				if len(pods.Items) > 1 {
					log.Info("More than one pod item found in node", node.Name)
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodCountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&podcountv1.PodCount{}).
		Complete(r)
}
