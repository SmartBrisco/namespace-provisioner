/*
Copyright 2026.

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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	platformv1alpha1 "github.com/SmartBrisco/namespace-provisioner/api/v1alpha1"
)

// ManagedNamespaceReconciler reconciles a ManagedNamespace object
type ManagedNamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=platform.platform.io,resources=managednamespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.platform.io,resources=managednamespaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platform.platform.io,resources=managednamespaces/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=resourcequotas,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch

func (r *ManagedNamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var managedNS platformv1alpha1.ManagedNamespace
	if err := r.Get(ctx, req.NamespacedName, &managedNS); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ManagedNamespace resource not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ManagedNamespace")
		return ctrl.Result{}, err
	}

	namespaceName := managedNS.Spec.Team + "-" + managedNS.Spec.Environment

	if err := r.reconcileNamespace(ctx, &managedNS, namespaceName); err != nil {
		log.Error(err, "Failed to reconcile namespace")
		return ctrl.Result{}, err
	}

	if err := r.reconcileResourceQuota(ctx, &managedNS, namespaceName); err != nil {
		log.Error(err, "Failed to reconcile resource quota")
		return ctrl.Result{}, err
	}

	if err := r.reconcileRBAC(ctx, &managedNS, namespaceName); err != nil {
		log.Error(err, "Failed to reconcile RBAC")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ManagedNamespaceReconciler) reconcileNamespace(ctx context.Context, managedNS *platformv1alpha1.ManagedNamespace, namespaceName string) error {
	ns := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: namespaceName}, ns)
	if apierrors.IsNotFound(err) {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
				Labels: map[string]string{
					"managed-by":  "namespace-provisioner",
					"team":        managedNS.Spec.Team,
					"environment": managedNS.Spec.Environment,
				},
			},
		}
		return r.Create(ctx, ns)
	}
	return err
}

func (r *ManagedNamespaceReconciler) reconcileResourceQuota(ctx context.Context, managedNS *platformv1alpha1.ManagedNamespace, namespaceName string) error {
	quota := &corev1.ResourceQuota{}
	err := r.Get(ctx, types.NamespacedName{Name: "default-quota", Namespace: namespaceName}, quota)
	if apierrors.IsNotFound(err) {
		quota = &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default-quota",
				Namespace: namespaceName,
			},
			Spec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{
					corev1.ResourceLimitsCPU:    resource.MustParse(managedNS.Spec.ResourceQuota.CPU),
					corev1.ResourceLimitsMemory: resource.MustParse(managedNS.Spec.ResourceQuota.Memory),
				},
			},
		}
		return r.Create(ctx, quota)
	}
	return err
}

func (r *ManagedNamespaceReconciler) reconcileRBAC(ctx context.Context, managedNS *platformv1alpha1.ManagedNamespace, namespaceName string) error {
	if err := r.reconcileRoleBinding(ctx, namespaceName, "admin-binding", "admin", managedNS.Spec.RBAC.Admins); err != nil {
		return err
	}
	if len(managedNS.Spec.RBAC.Viewers) > 0 {
		if err := r.reconcileRoleBinding(ctx, namespaceName, "viewer-binding", "view", managedNS.Spec.RBAC.Viewers); err != nil {
			return err
		}
	}
	return nil
}

func (r *ManagedNamespaceReconciler) reconcileRoleBinding(ctx context.Context, namespaceName, bindingName, roleName string, users []string) error {
	rb := &rbacv1.RoleBinding{}
	err := r.Get(ctx, types.NamespacedName{Name: bindingName, Namespace: namespaceName}, rb)
	if apierrors.IsNotFound(err) {
		subjects := make([]rbacv1.Subject, len(users))
		for i, user := range users {
			subjects[i] = rbacv1.Subject{
				Kind:     "User",
				Name:     user,
				APIGroup: "rbac.authorization.k8s.io",
			}
		}
		rb = &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bindingName,
				Namespace: namespaceName,
			},
			Subjects: subjects,
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     roleName,
			},
		}
		return r.Create(ctx, rb)
	}
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedNamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.ManagedNamespace{}).
		Named("managednamespace").
		Complete(r)
}

// Ensure fmt is used
var _ = fmt.Sprintf
