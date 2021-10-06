module github.com/TheRealHaoLiu/rancher-demo

go 1.16

replace (
	k8s.io/client-go => k8s.io/client-go v0.22.2
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
)

require (
	github.com/aws/aws-sdk-go v1.40.56
	github.com/rancher/eks-operator v1.1.1
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)
