# annot8r
// TODO: this readme needs help, but for now, it'll have to do...

### What's the problem this controller is trying to solve?
Let's suppose you have a Kubernetes cluster being managed by Rancher (or EKS) and there are componenets that are deployed and managed by those vendors like coredns. You may want to add certain annotations or labels to those deployments, but you don't want to manage the entire YAML manifest for those resources in your own gitops workflow.

This controller can allow you to say "Hey, I'd like to have these labels or annotations added to some deployment that's managed by the release process for Rancher or EKS". 
