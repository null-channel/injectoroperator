/*
Copyright 2021.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	applev1 "github.com/null-channel/apple-operator/api/v1"
	v1 "github.com/null-channel/apple-operator/api/v1"
)

// ContainerInjectorReconciler reconciles a ContainerInjector object
type ContainerInjectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	Selector = "injector"
)

//+kubebuilder:rbac:groups=apple.dev.thenullchannel,resources=containerinjectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apple.dev.thenullchannel,resources=containerinjectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apple.dev.thenullchannel,resources=containerinjectors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ContainerInjector object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ContainerInjectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	injector := &v1.ContainerInjector{}
	err := r.Client.Get(ctx, req.NamespacedName, injector)

	if err != nil {
		return ctrl.Result{}, err
	}

	// name of our custom finalizer
	myFinalizerName := "dev.thenullchannel.containerinjector/finalizer"

	// examine DeletionTimestamp to determine if object is under deletion
	if injector.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(injector.GetFinalizers(), myFinalizerName) {
			controllerutil.AddFinalizer(injector, myFinalizerName)
			if err := r.Update(ctx, injector); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(injector.GetFinalizers(), myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ctx, req.NamespacedName.Namespace, injector); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(injector, myFinalizerName)
			if err := r.Update(ctx, injector); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	deployments := &appsv1.DeploymentList{}
	err = r.Client.List(ctx, deployments, &client.ListOptions{Namespace: req.NamespacedName.Namespace})
	if err != nil {
		return ctrl.Result{}, err
	}

	deploys := []appsv1.Deployment{}
	for _, v := range deployments.Items {
		fmt.Println(v.Name)
		fmt.Println("Labels: ", v.ObjectMeta.Labels)
		if v.ObjectMeta.Labels[Selector] == injector.Spec.Label {
			deploys = append(deploys, v)
		}
	}

	for _, v := range deploys {
		println("deployment to update: ", v.Name)
		needsUpdated := true
		for _, c := range v.Spec.Template.Spec.Containers {
			if c.Name == Selector {
				needsUpdated = false
			}
		}

		if needsUpdated {
			newContainer := GetNewContainers(injector.Spec)
			v.Spec.Template.Spec.Containers = append(v.Spec.Template.Spec.Containers, newContainer)
			r.Client.Update(ctx, &v, &client.UpdateOptions{})
		}
	}

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

func (r *ContainerInjectorReconciler) deleteExternalResources(ctx context.Context, namespace string, containerInjector *v1.ContainerInjector) error {
	//
	// delete any external resources associated with the cronJob
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.

	deployments := &appsv1.DeploymentList{}
	err := r.Client.List(ctx, deployments, &client.ListOptions{Namespace: namespace})

	if err != nil {
		return err
	}

	deploys := []appsv1.Deployment{}
	for _, v := range deployments.Items {
		fmt.Println(v.Name)
		fmt.Println("Labels: ", v.ObjectMeta.Labels)

		for _, container := range v.Spec.Template.Spec.Containers {
			if container.Name == Selector {
				deploys = append(deploys, v)
			}
		}
	}

	for _, v := range deploys {
		println("deployment to remove: ", v.Name)

		newContaienrs := []corev1.Container{}

		for _, container := range v.Spec.Template.Spec.Containers {
			if container.Name != Selector {
				newContaienrs = append(newContaienrs, container)
			}
		}

		v.Spec.Template.Spec.Containers = newContaienrs

		r.Client.Update(ctx, &v, &client.UpdateOptions{})

	}

	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func GetNewContainers(spec v1.ContainerInjectorSpec) corev1.Container {
	return corev1.Container{
		Name:    Selector,
		Image:   spec.Image,
		Command: spec.Command,
		Args:    spec.Args,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ContainerInjectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&applev1.ContainerInjector{}).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			handler.EnqueueRequestsFromMapFunc(r.GetAll),
		).
		Complete(r)
}

func (r *ContainerInjectorReconciler) GetAll(o client.Object) []ctrl.Request {
	result := []ctrl.Request{}

	injectorList := v1.ContainerInjectorList{}
	r.Client.List(context.Background(), &injectorList)

	for _, labeler := range injectorList.Items {
		result = append(result, ctrl.Request{NamespacedName: client.ObjectKey{Namespace: labeler.Namespace, Name: labeler.Name}})
	}

	return result
}
