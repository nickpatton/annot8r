package annoLabler

import (
	appsv1 "k8s.io/api/apps/v1"
)

func AnnotateDeployment(name string, namespace string, annotations map[string]string) appsv1.Deployment {
	var someDeployment appsv1.Deployment
	someDeployment.Name = name
	someDeployment.Namespace = namespace
	someDeployment.Spec.Template.ObjectMeta.Annotations = annotations

	return someDeployment
}
