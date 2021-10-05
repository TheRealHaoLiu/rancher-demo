# rancher-demo
## eks demo scenarios 

on a OpenShift cluster I am able to deploy rancker eks-operator (bonus point via helm)

in a contained "cluster namespace" I am able to create a "credential secret" and triger a basic cluster create

I am able to perform day 2 operation with eks-operator
- scaling up worker nodes
- upgrade kube version 

I am able to fully cleanup the delete the cluster

check lists
[ ] cluster create
[ ] cluster scale up/down
[ ] cluster delete
[ ] no "not build from source" binary is used

## hacks
Generate user token for the eks cluster
`go run hack/aws-generate-eks-token.go CLUSTER_NAME`
