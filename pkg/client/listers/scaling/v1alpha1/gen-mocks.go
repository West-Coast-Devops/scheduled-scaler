package v1alpha1

//go:generate mockgen -destination=mock_$GOPACKAGE/scheduledscaler.go k8s.restdev.com/operators/pkg/client/listers/scaling/v1alpha1 ScheduledScalerLister,ScheduledScalerNamespaceLister
