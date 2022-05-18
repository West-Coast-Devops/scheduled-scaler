package v1alpha1

//go:generate mockgen -destination=mock_$GOPACKAGE/scaling_client.go k8s.restdev.com/operators/pkg/client/clientset/versioned/typed/scaling/v1alpha1 ScalingV1alpha1Interface
//go:generate mockgen -destination=mock_$GOPACKAGE/scheduledscaler.go k8s.restdev.com/operators/pkg/client/clientset/versioned/typed/scaling/v1alpha1 ScheduledScalersGetter,ScheduledScalerInterface
