package k8s

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
	authenticationv1 "k8s.io/api/authentication/v1"
)
func CreateRemoteTokenForUser(edgeClusterName string, edgeClusterPort int, tokenSecretName string) (string, error) {
	// Retrieve the access token
	accessToken, err := getAccessTokenFromTokenReview(tokenSecretName)
	if err != nil {
		return "", err
	}

	// Prepare the remote cluster API URL
	apiURL := fmt.Sprintf("https://%s:%d", edgeClusterName, edgeClusterPort)

	// Create a Kubernetes client for the remote cluster
	client, err := createKubernetesClient(apiURL, accessToken)
	if err != nil {
		return "", err
	}

	// Specify the user, namespace, and token expiration
	username := edgeClusterName + "-cp-access"
	expiresIn := time.Hour * 24 // 24-hour token expiration
	namespace := "default"

	// Create a ServiceAccountTokenRequest
	tokenRequest := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences:         []string{"api"}, // Define the intended audiences for the token
			ExpirationSeconds: int64(expiresIn.Seconds()),
		},
	}

	// Create the token using the TokenRequest API
	result, err := client.CoreV1().ServiceAccounts(namespace).CreateToken(context.Background(), username, tokenRequest, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create token: %v", err)
	}

	return result.Status.Token, nil
}
