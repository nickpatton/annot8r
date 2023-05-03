package annoLabler

import (
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Add annotations to deployment's pod template spec, remove them if 'remove' is true
func AnnotateDeploymentPodSpec(annotations map[string]string, deployment *appsv1.Deployment, remove bool) {
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{
			"annot8r.kube.tools/v1": "true",
		}
	}

	if remove {
		log.Log.Info("removing annotations from deployment: " + deployment.Name)
		for key, _ := range annotations {
			delete(deployment.Spec.Template.Annotations, key)
		}
		delete(deployment.Spec.Template.Annotations, "annot8r.kube.tools/v1")
	} else {
		log.Log.Info("adding annotations to deployment: " + deployment.Name)
		maps.Copy(deployment.Spec.Template.ObjectMeta.Annotations, annotations)
	}
}
