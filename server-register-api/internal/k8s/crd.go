package k8s

import (
	"context"
	"log"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func EnsureEdgeClusterInfoCRD() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		return err
	}

	crdName := "edgeclusterinfos.xddevelopment.com"
	_, err = clientset.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), crdName, metav1.GetOptions{})
	if err == nil {
		log.Printf("EdgeClusterInfo CRD already exists")
		return nil
	}

	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "xddevelopment.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "EdgeClusterInfo",
				ListKind: "EdgeClusterInfoList",
				Plural:   "edgeclusterinfos",
				Singular: "edgeclusterinfo",
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Storage: true,
					Served:  true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"port": {
											Type:    "integer",
											Minimum: &[]float64{10000}[0],
											Maximum: &[]float64{30000}[0],
										},
										"expose_name": {
											Type:   "string",
											Format: "^[a-z]{10}$",
										},
									},
									Required: []string{"port", "expose_name"},
								},
								"metadata": {
									Type: "object",
								},
							},
						},
					},

				},
			},
		},
	}

	_, err = clientset.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), crd, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	log.Printf("EdgeClusterInfo CRD created")
	return nil
}
