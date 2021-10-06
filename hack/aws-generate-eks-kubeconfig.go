package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	eksv1 "github.com/rancher/eks-operator/pkg/apis/eks.cattle.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

//input namespace/name of `eksclusterconfig.eks.cattle.io`
// - check if eksclusterconfig.phase == active else exit
// - get secret in eksclusterconfig.metadata.namespace with the same name as eksclusterconfig.metadata.name
// - generate base kubeconfig from template using secret.data.ca and secret.data.endpoint
// - from eksclusterconfig.spec.amazonCredentialSecret locate credential secret in format "namespace:name"
// - intialize sts client with the credential in secret to get eks token
/*
	stsClient := stsiface.STSAPI(sts.New(session.New()))

	req, _ := stsClient.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	req.HTTPRequest.Header.Add("x-k8s-aws-id", "hao-hack-rancher-eks")

	presignedURL, _ := req.Presign(15 * time.Minute)

	encodedURL := base64.RawURLEncoding.EncodeToString([]byte(presignedURL))

	fmt.Printf("%s.%s", "k8s-aws-v1", encodedURL)
*/
// - use eks token and generate the functioning kubeconfig

func main() {

	clusterName := os.Args[1]

	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create Dynamic Clientset
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// get eksclusterconfig
	eksGVR := schema.GroupVersionResource{Group: "eks.cattle.io", Version: "v1", Resource: "eksclusterconfigs"}
	eksUnstructured, err := dynClient.Resource(eksGVR).Namespace(clusterName).Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	eks := &eksv1.EKSClusterConfig{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(eksUnstructured.UnstructuredContent(), eks)
	if err != nil {
		panic(err.Error())
	}

	// check cluster finish provision
	if eks.Status.Phase != "active" {
		panic(fmt.Errorf("cluster not ready yet"))
	}

	secretGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}

	//get cluster secret
	clusterSecretUnstructured, err := dynClient.Resource(secretGVR).Namespace(clusterName).Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	clusterSecret := &corev1.Secret{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(clusterSecretUnstructured.UnstructuredContent(), clusterSecret)
	if err != nil {
		panic(err.Error())
	}

	//get cloud credential secret
	awsSecretNsN := strings.Split(eks.Spec.AmazonCredentialSecret, ":")
	var awsSecretNamespace string
	var awsSecretName string
	if len(awsSecretNsN) == 1 {
		awsSecretNamespace = clusterName
		awsSecretName = awsSecretNsN[0]
	} else if len(awsSecretNsN) == 2 {
		awsSecretNamespace = awsSecretNsN[0]
		awsSecretName = awsSecretNsN[1]
	}

	awsSecretUnstructured, err := dynClient.Resource(secretGVR).Namespace(awsSecretNamespace).Get(context.TODO(), awsSecretName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	awsSecret := &corev1.Secret{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(awsSecretUnstructured.UnstructuredContent(), awsSecret)
	if err != nil {
		panic(err.Error())
	}

	// fmt.Println(eks)
	// fmt.Println(clusterSecret)
	// fmt.Println(awsSecret)

	eksToken := getEKSToken(clusterName)
	// fmt.Println(eksToken)

	caData, err := base64.StdEncoding.DecodeString(string(clusterSecret.Data["ca"]))
	if err != nil {
		panic(err.Error())
	}

	//generate kubeconfig for the eks cluster
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters[eks.Name] = &clientcmdapi.Cluster{
		Server:                   string(clusterSecret.Data["endpoint"]),
		CertificateAuthorityData: caData,
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts[eks.Name] = &clientcmdapi.Context{
		Cluster:  eks.Name,
		AuthInfo: eks.Name,
	}

	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos[eks.Name] = &clientcmdapi.AuthInfo{
		Token: string(eksToken),
	}

	eksClientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: eks.Name,
		AuthInfos:      authinfos,
	}

	eksKubeConfig, err := clientcmd.Write(eksClientConfig)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(eksKubeConfig))
}

func getEKSToken(clusterName string) string {
	//TODO: use cred in AWS secret instead of env var
	stsClient := stsiface.STSAPI(sts.New(session.New()))

	req, _ := stsClient.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	req.HTTPRequest.Header.Add("x-k8s-aws-id", clusterName)

	presignedURL, _ := req.Presign(15 * time.Minute)

	encodedURL := base64.RawURLEncoding.EncodeToString([]byte(presignedURL))

	return fmt.Sprintf("%s.%s", "k8s-aws-v1", encodedURL)
}

//output kubeconfig
