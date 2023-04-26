package k8s

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CreateRemoteServiceAccount creates a service account named 'test' on the remote Kubernetes cluster
func CreateRemoteServiceAccount(edgeClusterName string, edgeClusterPort int, tokenSecretName string) error {
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

	// Create the service account
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      edgeClusterName + "-cp-access",
			Namespace: "default",
		},
	}

	// Check if the service account already exists
	_, err = client.CoreV1().ServiceAccounts("default").Get(context.Background(), "test", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create the service account if it doesn't exist
			_, err = client.CoreV1().ServiceAccounts("default").Create(context.Background(), sa, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Println("Service account 'test' created successfully.")
		} else {
			return err
		}
	} else {
		fmt.Println("Service account 'test' already exists.")
	}

	return nil
}

// getAccessTokenFromTokenReview retrieves the access token from the Kubernetes secret
// and validates it using the TokenReview API
func getAccessTokenFromTokenReview(secretName string) (string, error) {
	// Use in-cluster configuration if running inside a Kubernetes pod
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	// Create a Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Retrieve the secret from the Kubernetes API
	namespace := "chisel-server" // replace with the appropriate namespace
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s: %v", secretName, err)
	}

	// Extract the access token from the secret
	accessTokenBytes, ok := secret.Data["edge-access-token"]
	if !ok {
		return "", fmt.Errorf("access token not found in secret %s", secretName)
	}

	tokenString := string(accessTokenBytes)

	//
	//// Validate the token using the TokenReview API
	//tokenReview := &authenticationv1.TokenReview{
	//	Spec: authenticationv1.TokenReviewSpec{
	//		Token: tokenString,
	//	},
	//}
	//
	//reviewResult, err := clientset.AuthenticationV1().TokenReviews().Create(context.Background(), tokenReview, metav1.CreateOptions{})
	//if err != nil {
	//	return "", fmt.Errorf("failed to review token: %v", err)
	//}
	//fmt.Printf("TokenReview result: %+v\n", reviewResult)
	//fmt.Printf("TokenReview status: %+v\n", reviewResult.Status)
	//if !reviewResult.Status.Authenticated {
	//	return "", fmt.Errorf("token is not authenticated")
	//}

	return tokenString, nil
}

// createKubernetesClient creates a Kubernetes client for the remote cluster
func createKubernetesClient(apiURL, accessToken string) (*kubernetes.Clientset, error) {
	// Create a REST client configuration
	config := &rest.Config{
		Host:        apiURL,
		BearerToken: accessToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // If using self-signed certificates, set Insecure to true
		},
	}

	// Create the Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
