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

// simple program to demo how to get EKS user token
func main() {
	clusterName := os.Args[1]

	stsClient := stsiface.STSAPI(sts.New(session.New()))

	req, _ := stsClient.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	req.HTTPRequest.Header.Add("x-k8s-aws-id", clusterName)

	presignedURL, _ := req.Presign(15 * time.Minute)

	encodedURL := base64.RawURLEncoding.EncodeToString([]byte(presignedURL))

	fmt.Printf("%s.%s", "k8s-aws-v1", encodedURL)
}
