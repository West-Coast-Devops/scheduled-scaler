package step

import scalingv1alpha1 "k8s.restdev.com/operators/pkg/apis/scaling/v1alpha1"

func Parse(step scalingv1alpha1.ScheduledScalerStep) (min, max *int32) {
	if step.Mode == "range" {
		min = step.MinReplicas
		max = step.MaxReplicas
	}

	if step.Mode == "fixed" {
		min = step.Replicas
		max = step.Replicas
	}

	return
}
