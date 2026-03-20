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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	platformv1alpha1 "github.com/SmartBrisco/namespace-provisioner/api/v1alpha1"
)

var _ = Describe("ManagedNamespace Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		managednamespace := &platformv1alpha1.ManagedNamespace{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ManagedNamespace")
			err := k8sClient.Get(ctx, typeNamespacedName, managednamespace)
			if err != nil && errors.IsNotFound(err) {
				resource := &platformv1alpha1.ManagedNamespace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: platformv1alpha1.ManagedNamespaceSpec{
						Team:        "payments",
						Environment: "dev",
						ResourceQuota: platformv1alpha1.ResourceQuotaSpec{
							CPU:    "2",
							Memory: "4Gi",
						},
						RBAC: platformv1alpha1.RBACSpec{
							Admins:  []string{"alice"},
							Viewers: []string{"bob"},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &platformv1alpha1.ManagedNamespace{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance ManagedNamespace")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ManagedNamespaceReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create the namespace", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ManagedNamespaceReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the namespace was created")
			ns := &corev1.Namespace{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "payments-dev"}, ns)
			Expect(err).NotTo(HaveOccurred())
			Expect(ns.Labels["team"]).To(Equal("payments"))
			Expect(ns.Labels["environment"]).To(Equal("dev"))
		})

		It("should create the resource quota", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ManagedNamespaceReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the resource quota was created")
			quota := &corev1.ResourceQuota{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "default-quota", Namespace: "payments-dev"}, quota)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create admin role binding", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ManagedNamespaceReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the admin role binding was created")
			rb := &rbacv1.RoleBinding{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "admin-binding", Namespace: "payments-dev"}, rb)
			Expect(err).NotTo(HaveOccurred())
			Expect(rb.Subjects[0].Name).To(Equal("alice"))
		})
	})
})
