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

package controllers

import (
	"context"
	"fmt"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	// "annot8r/annoLabler"

	kubev1 "annot8r/api/v1"

	appsv1 "k8s.io/api/apps/v1"
)

// Annot8rReconciler reconciles a Annot8r object
type Annot8rReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kube.tools,resources=annot8rs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kube.tools,resources=annot8rs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kube.tools,resources=annot8rs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Annot8r object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *Annot8rReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var annot8r kubev1.Annot8r
	if err := r.Get(ctx, req.NamespacedName, &annot8r); err != nil {
		log.Log.Error(err, "unable to fetch Annot8r")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// name of our custom finalizer
	finalizerName := "annot8r.kube.tools/finalizer"

	// examine DeletionTimestamp to determine if object is under deletion
	if annot8r.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(&annot8r, finalizerName) {
			controllerutil.AddFinalizer(&annot8r, finalizerName)
			if err := r.Update(ctx, &annot8r); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(&annot8r, finalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.unDoAnnotationsLabels(&annot8r, ctx); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&annot8r, finalizerName)
			if err := r.Update(ctx, &annot8r); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	switch annot8r.Spec.Kind {
	case "deployment":
		log.Log.Info("processing annotations and labels on deployment")
		// desiredDeployment = annoLabler.AnnotateDeployment(annot8r.Spec.Name, annot8r.Spec.Namespace, annot8r.Spec.Annotations)

		var currentDeployment appsv1.Deployment
		var currentDeploymentLookup types.NamespacedName
		currentDeploymentLookup.Name = annot8r.Spec.Name
		currentDeploymentLookup.Namespace = annot8r.Spec.Namespace
		if err := r.Get(ctx, currentDeploymentLookup, &currentDeployment); err != nil {
			log.Log.Error(err, "unable to lookup the deployment")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		if currentDeployment.Spec.Template.ObjectMeta.Annotations == nil {
			currentDeployment.Spec.Template.ObjectMeta.Annotations = map[string]string{
				"annot8r.kube.tools/v1": "true",
			}
		}
		maps.Copy(currentDeployment.Spec.Template.ObjectMeta.Annotations, annot8r.Spec.Annotations)
		if err := r.Update(ctx, &currentDeployment); err != nil {
			log.Log.Error(err, "unable to update the deployment")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

	default:
		fmt.Println("Error.")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Annot8rReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubev1.Annot8r{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				// The reconciler adds a finalizer so we perform clean-up
				// when the delete timestamp is added
				// Suppress Delete events to avoid filtering them out in the Reconcile function
				log.Log.Info("deleting " + e.Object.GetName() + " Annot8r resource")
				return false
			},
		}).
		Complete(r)
}

func (r *Annot8rReconciler) unDoAnnotationsLabels(annot8r *kubev1.Annot8r, ctx context.Context) error {
	log.Log.Info("removing annotations and/or labels from resource")
	switch annot8r.Spec.Kind {
	case "deployment":
		var currentDeployment appsv1.Deployment
		var currentDeploymentLookup types.NamespacedName
		currentDeploymentLookup.Name = annot8r.Spec.Name
		currentDeploymentLookup.Namespace = annot8r.Spec.Namespace
		if err := r.Get(ctx, currentDeploymentLookup, &currentDeployment); err != nil {
			log.Log.Error(err, "unable to lookup the deployment")
			return client.IgnoreNotFound(err)
		}
		for key, _ := range annot8r.Spec.Annotations {
			delete(currentDeployment.Spec.Template.ObjectMeta.Annotations, key)
		}

		// remove default annot8r annotation (added in cases where no original annotations exist on the resource)
		delete(currentDeployment.Spec.Template.ObjectMeta.Annotations, "annot8r.kube.tools/v1")

		if err := r.Update(ctx, &currentDeployment); err != nil {
			log.Log.Error(err, "unable to update the deployment")
			return client.IgnoreNotFound(err)
		}
	default:
		fmt.Println("Error undoing annotations.")
	}
	return nil
}
