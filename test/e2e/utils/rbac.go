/*
Copyright 2018-2022 The Kubernetes Authors.

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

package utils

import (
	"context"
	"flag"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

var (
	openShift = flag.Bool("nfd.openshift", false, "Enable OpenShift specific bits")
)

// ConfigureRBAC creates required RBAC configuration
func ConfigureRBAC(ctx context.Context, cs clientset.Interface, ns string) error {
	_, err := createServiceAccount(ctx, cs, "nfd-master-e2e", ns)
	if err != nil {
		return err
	}

	_, err = createServiceAccount(ctx, cs, "nfd-worker-e2e", ns)
	if err != nil {
		return err
	}

	_, err = createServiceAccount(ctx, cs, "nfd-topology-updater-e2e", ns)
	if err != nil {
		return err
	}

	_, err = createClusterRoleMaster(ctx, cs)
	if err != nil {
		return err
	}

	_, err = createRoleWorker(ctx, cs, ns)
	if err != nil {
		return err
	}

	_, err = createClusterRoleTopologyUpdater(ctx, cs)
	if err != nil {
		return err
	}

	_, err = createClusterRoleBindingMaster(ctx, cs, ns)
	if err != nil {
		return err
	}

	_, err = createRoleBindingWorker(ctx, cs, ns)
	if err != nil {
		return err
	}

	_, err = createClusterRoleBindingTopologyUpdater(ctx, cs, ns)
	if err != nil {
		return err
	}

	return nil
}

// DeconfigureRBAC removes RBAC configuration
func DeconfigureRBAC(ctx context.Context, cs clientset.Interface, ns string) error {
	err := cs.RbacV1().ClusterRoleBindings().Delete(ctx, "nfd-topology-updater-e2e", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = cs.RbacV1().ClusterRoleBindings().Delete(ctx, "nfd-master-e2e", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = cs.RbacV1().RoleBindings(ns).Delete(ctx, "nfd-worker-e2e", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = cs.RbacV1().ClusterRoles().Delete(ctx, "nfd-topology-updater-e2e", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = cs.RbacV1().ClusterRoles().Delete(ctx, "nfd-master-e2e", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = cs.RbacV1().Roles(ns).Delete(ctx, "nfd-worker-e2e", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = cs.CoreV1().ServiceAccounts(ns).Delete(ctx, "nfd-topology-updater-e2e", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = cs.CoreV1().ServiceAccounts(ns).Delete(ctx, "nfd-master-e2e", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = cs.CoreV1().ServiceAccounts(ns).Delete(ctx, "nfd-worker-e2e", metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// Configure service account
func createServiceAccount(ctx context.Context, cs clientset.Interface, name, ns string) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	return cs.CoreV1().ServiceAccounts(ns).Create(ctx, sa, metav1.CreateOptions{})
}

// Configure cluster role required by NFD Master
func createClusterRoleMaster(ctx context.Context, cs clientset.Interface) (*rbacv1.ClusterRole, error) {
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nfd-master-e2e",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes", "nodes/status"},
				Verbs:     []string{"get", "list", "patch", "update"},
			},
			{
				APIGroups: []string{"nfd.k8s-sigs.io"},
				Resources: []string{"nodefeatures", "nodefeaturerules"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	if *openShift {
		cr.Rules = append(cr.Rules,
			rbacv1.PolicyRule{
				// needed on OpenShift clusters
				APIGroups:     []string{"security.openshift.io"},
				Resources:     []string{"securitycontextconstraints"},
				ResourceNames: []string{"hostaccess"},
				Verbs:         []string{"use"},
			})
	}
	return cs.RbacV1().ClusterRoles().Update(ctx, cr, metav1.UpdateOptions{})
}

// Configure role required by NFD Worker
func createRoleWorker(ctx context.Context, cs clientset.Interface, ns string) (*rbacv1.Role, error) {
	cr := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nfd-worker-e2e",
			Namespace: ns,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"nfd.k8s-sigs.io"},
				Resources: []string{"nodefeatures"},
				Verbs:     []string{"create", "get", "update"},
			},
		},
	}
	return cs.RbacV1().Roles(ns).Update(ctx, cr, metav1.UpdateOptions{})
}

// Configure cluster role required by NFD Topology Updater
func createClusterRoleTopologyUpdater(ctx context.Context, cs clientset.Interface) (*rbacv1.ClusterRole, error) {
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nfd-topology-updater-e2e",
		},
		// the Topology Updater doesn't need to access any kube object:
		// it reads from the podresources socket and it sends updates to the
		// nfd-master using the gRPC interface.
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"topology.node.k8s.io"},
				Resources: []string{"noderesourcetopologies"},
				Verbs: []string{
					"create",
					"get",
					"update",
				},
			},
		},
	}
	if *openShift {
		cr.Rules = append(cr.Rules,
			rbacv1.PolicyRule{
				// needed on OpenShift clusters
				APIGroups:     []string{"security.openshift.io"},
				Resources:     []string{"securitycontextconstraints"},
				ResourceNames: []string{"hostaccess"},
				Verbs:         []string{"use"},
			})
	}
	return cs.RbacV1().ClusterRoles().Update(ctx, cr, metav1.UpdateOptions{})
}

// Configure cluster role binding required by NFD Master
func createClusterRoleBindingMaster(ctx context.Context, cs clientset.Interface, ns string) (*rbacv1.ClusterRoleBinding, error) {
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nfd-master-e2e",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "nfd-master-e2e",
				Namespace: ns,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "nfd-master-e2e",
		},
	}

	return cs.RbacV1().ClusterRoleBindings().Update(ctx, crb, metav1.UpdateOptions{})
}

// Configure role binding required by NFD Master
func createRoleBindingWorker(ctx context.Context, cs clientset.Interface, ns string) (*rbacv1.RoleBinding, error) {
	crb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nfd-worker-e2e",
			Namespace: ns,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "nfd-worker-e2e",
				Namespace: ns,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     "nfd-worker-e2e",
		},
	}

	return cs.RbacV1().RoleBindings(ns).Update(ctx, crb, metav1.UpdateOptions{})
}

// Configure cluster role binding required by NFD Topology Updater
func createClusterRoleBindingTopologyUpdater(ctx context.Context, cs clientset.Interface, ns string) (*rbacv1.ClusterRoleBinding, error) {
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nfd-topology-updater-e2e",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "nfd-topology-updater-e2e",
				Namespace: ns,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "nfd-topology-updater-e2e",
		},
	}

	return cs.RbacV1().ClusterRoleBindings().Update(ctx, crb, metav1.UpdateOptions{})
}
