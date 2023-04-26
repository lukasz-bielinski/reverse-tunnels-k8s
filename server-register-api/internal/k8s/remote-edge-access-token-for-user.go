package k8s

import (
	"context"
	"fmt"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
	"io/ioutil"
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
	serviceAccountName := edgeClusterName + "-cp-access"
	expiresIn := time.Minute * 60 // 1-hour token expiration
	namespace := "default"


	// Create a new TokenRequest for the service account
	tokenRequest := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences:         []string{"https://kubernetes.default.svc.cluster.local"},
			ExpirationSeconds: int64Ptr(int64(expiresIn.Seconds())), // Set the token's expiration time (in seconds)
		},
	}

	token, err := client.CoreV1().ServiceAccounts(namespace).CreateToken(context.Background(), serviceAccountName, tokenRequest, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create token: %v", err)
	}
	fmt.Println("Token:", token.Status.Token)
	return token.Status.Token, nil

}

func int64Ptr(i int64) *int64 {
	return &i
}

func GenerateKubeConfig(clusterName string, chiselTunnelDomain string, exposeName string, token string) string {
	kubeConfigTemplate := `
apiVersion: v1
kind: Config
clusters:
- name: %s
  cluster:
    server: %s
    insecure-skip-tls-verify: true
contexts:
- name: %s
  context:
    cluster: %s
    user: %s
current-context: %s
users:
- name: %s
  user:
    token: %s
`

	serverURL := fmt.Sprintf("https://%s/%s/", chiselTunnelDomain, exposeName)

	kubeConfig := fmt.Sprintf(
		kubeConfigTemplate,
		clusterName,
		serverURL,
		clusterName,
		clusterName,
		clusterName,
		clusterName,
		clusterName,
		token,
	)

	return kubeConfig
}

func WriteKubeConfigToFile(kubeConfigContent string, edgeClusterName string) error {
	filePath := fmt.Sprintf("/data/%s-kubeconfig.yaml", edgeClusterName)
	err := ioutil.WriteFile(filePath, []byte(kubeConfigContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to file: %v", err)
	}
	return nil
}