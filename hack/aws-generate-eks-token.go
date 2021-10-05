package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

const CLUSTER_NAME_HEADER = "x-k8s-aws-id"

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

	stsClient := stsiface.STSAPI(sts.New(session.New()))

	req, _ := stsClient.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	req.HTTPRequest.Header.Add("x-k8s-aws-id", clusterName)

	presignedURL, _ := req.Presign(15 * time.Minute)

	encodedURL := base64.RawURLEncoding.EncodeToString([]byte(presignedURL))

	fmt.Printf("%s.%s", "k8s-aws-v1", encodedURL)
}

//output kubeconfig
