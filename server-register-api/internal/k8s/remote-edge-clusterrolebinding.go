package k8s

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateRemoteServiceAccount creates a service account named 'test' on the remote Kubernetes cluster
func CreateRemoteClusterRoleBinding(edgeClusterName string, edgeClusterPort int, tokenSecretName string) error {
	// Retrieve the access token
	accessToken, err := getAccessTokenFromTokenReview(tokenSecretName)
	if err != nil {
		return err
	}

	// Prepare the remote cluster API URL
	apiURL := fmt.Sprintf("https://%s:%d", edgeClusterName, edgeClusterPort)

	// Create a Kubernetes client for the remote cluster
	client, err := createKubernetesClient(apiURL, accessToken)
	if err != nil {
		return err
	}

	// Check if the ClusterRoleBinding already exists
	_, err = client.RbacV1().ClusterRoleBindings().Get(context.Background(), edgeClusterName+"-cp-access", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create the ClusterRoleBinding if it doesn't exist
			clusterRoleBinding := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: edgeClusterName + "-cp-access",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     "cluster-admin", // Replace with the desired ClusterRole name
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      edgeClusterName + "-cp-access",
						Namespace: "default",
					},
				},
			}

			_, err = client.RbacV1().ClusterRoleBindings().Create(context.Background(), clusterRoleBinding, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Println("ClusterRoleBinding '" + edgeClusterName + "-cp-access' created successfully.")
		} else {
			return err
		}
	} else {
		fmt.Println("ClusterRoleBinding '" + edgeClusterName + "-cp-access' already exists.")
	}
	return nil
}
