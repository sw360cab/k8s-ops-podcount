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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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
	log.Info("reconciling custom resource")

	// Get the PodCount resource
	var podCount podcountv1.PodCount
	var podCountList podcountv1.PodCountList

	if err := r.List(ctx, &podCountList); err != nil || len(podCountList.Items) != 1 {
		log.Error(err, "unable to fetch CRD of counter")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	podCount = podCountList.Items[0]

	// Get all the nodes
	var nodes corev1.NodeList
	var pods corev1.PodList
	if err := r.List(ctx, &nodes); err != nil {
		log.Error(err, "unable to list nodes")
	} else {
		for _, node := range nodes.Items {
			opts := []client.ListOption{
				client.InNamespace(req.NamespacedName.Namespace),
				client.MatchingLabels{"name": podCount.Spec.Name},
				client.MatchingFields{"spec.nodeName": node.Name},
			}

			// get pods in node
			if err := r.List(ctx, &pods, opts...); err != nil {
				log.Error(err, "unable to list pods")
			} else {
				if len(pods.Items) > podCount.Spec.PodsPerNode || podCount.Status.MaxPodsExceeded {
					log.Info("Ko -> Node exceeded the maximum number of Pods per node ",
						"name", node.Name, "spec", podCount.Spec)
					// Update podCount status
					podCount.Status.MaxPodsExceeded = true
					return ctrl.Result{}, nil
				} else if len(pods.Items) == podCount.Spec.PodsPerNode {
					log.Info("OK -> Node has reached the maximum number of Pods",
						"name", node.Name, "spec", podCount.Spec)
				}
				podCount.Status.MaxPodsExceeded = false
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodCountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// create index on node name
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "spec.nodeName",
		func(rawObj client.Object) []string {
			pod := rawObj.(*corev1.Pod)
			return []string{pod.Spec.NodeName}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&podcountv1.PodCount{}).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.mapPodsReqToPodCountReq),
		).
		// filters
		// WithEventFilter(predicate.Funcs{
		// 	UpdateFunc: func(e event.UpdateEvent) bool { return false },
		// 	DeleteFunc: func(e event.DeleteEvent) bool { return false },
		// }).
		Complete(r)
}

func (r *PodCountReconciler) mapPodsReqToPodCountReq(ctx context.Context, pod client.Object) []reconcile.Request {
	var pods corev1.PodList
	req := []reconcile.Request{}
	log := log.FromContext(ctx)

	// Get the PodCount resource
	var podCount podcountv1.PodCount
	var podCountList podcountv1.PodCountList
	if err := r.List(ctx, &podCountList); err != nil || len(podCountList.Items) != 1 {
		log.Error(err, "unable to fetch CRD of counter")
		return req
	}
	podCount = podCountList.Items[0]
	opts := []client.ListOption{
		client.MatchingLabels{"name": podCount.Spec.Name},
	}

	// get pods in node
	if err := r.List(ctx, &pods, opts...); err != nil {
		log.Error(err, "unable to list pods")
	} else {
		for _, item := range pods.Items {
			if item.Name == pod.GetName() {
				log.Info("Pod Ready", "item", item.Name, "name", pod.GetName())
				req = append(req, reconcile.Request{
					NamespacedName: types.NamespacedName{Name: item.Name, Namespace: item.Namespace},
				})
				log.Info("pod custom resource issued an event", "name", pod.GetName())
			}
		}
	}

	return req
}
