package mock_external

//go:generate mockgen -destination=kubernetes.go -package mock_external k8s.io/client-go/kubernetes Interface
//go:generate mockgen -destination=corev1.go -package mock_external k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface,ConfigMapInterface
//go:generate mockgen -destination=autoscalingv1.go -package mock_external k8s.io/client-go/kubernetes/typed/autoscaling/v1 AutoscalingV1Interface,HorizontalPodAutoscalerInterface
