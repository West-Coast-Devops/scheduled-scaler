package main

import (
	"flag"
	"fmt"
	"go.uber.org/multierr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"reflect"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/robfig/cron"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	scalingv1alpha1 "k8s.restdev.com/operators/pkg/apis/scaling/v1alpha1"
	clientset "k8s.restdev.com/operators/pkg/client/clientset/versioned"
	informers "k8s.restdev.com/operators/pkg/client/informers/externalversions"
	listers "k8s.restdev.com/operators/pkg/client/listers/scaling/v1alpha1"
	scalingcron "k8s.restdev.com/operators/pkg/services/scaling/cron"
	scalingmetadata "k8s.restdev.com/operators/pkg/services/scaling/metadata"
	scalingstep "k8s.restdev.com/operators/pkg/services/scaling/step"
)

/**
 *
 * Target Struct
 *
 */
type ScheduledScalerTarget struct {
	Name string
	Kind string
	Cron *cron.Cron
}

/**
 *
 * Scaling Controller Struct
 *
 */
type ScheduledScalerController struct {
	maxRetries             int
	cronProxy              scalingcron.CronProxy
	informerFactory        informers.SharedInformerFactory
	restdevClient          clientset.Interface
	kubeClient             kubernetes.Interface
	scheduledScalersLister listers.ScheduledScalerLister
	scheduledScalersSynced cache.InformerSynced
	scheduledScalerTargets []ScheduledScalerTarget
}

/**
 *
 * Start the shared informers and wait for cache to sync
 *
 */
func (c *ScheduledScalerController) Run(stopCh chan struct{}) error {
	c.informerFactory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, c.scheduledScalersSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

// validateScheduledScaler validates the scheduledScaler
func (c *ScheduledScalerController) validateScheduledScaler(scheduledScaler *scalingv1alpha1.ScheduledScaler) error {
	if scheduledScaler == nil {
		return fmt.Errorf("scheduledScaler is nil")
	}
	var err error
	for _, step := range scheduledScaler.Spec.Steps {
		schedule, stepErr := c.cronProxy.Parse(step.Runat)
		if stepErr != nil {
			err = multierr.Append(err, fmt.Errorf("error parsing Runat %s: %w", step.Runat, stepErr))
			continue
		}
		when := schedule.Next(time.Now())
		if when.IsZero() {
			err = multierr.Append(err, fmt.Errorf("invalid Runat %s; will never fire", step.Runat))
		}
	}
	return err
}

/**
 *
 * Add methods
 *
 * These methods will handle new resources
 *
 */
func (c *ScheduledScalerController) scheduledScalerAdd(obj interface{}) {
	scheduledScaler, ok := obj.(*scalingv1alpha1.ScheduledScaler)
	if !ok {
		utilruntime.HandleError(fmt.Errorf(
			"object %T is not a *scalingv1alpha1.ScheduledScaler; will not add", obj))
		return
	}
	if err := c.validateScheduledScaler(scheduledScaler); err != nil {
		utilruntime.HandleError(fmt.Errorf(
			"error validating scheduledScaler %#v: %w; will not add", scheduledScaler, err))
		return
	}
	if scheduledScaler.Spec.Target.Kind == "HorizontalPodAutoscaler" {
		c.scheduledScalerHpaCronAdd(scheduledScaler)
	} else if scheduledScaler.Spec.Target.Kind == "InstanceGroup" {
		c.scheduledScalerIgCronAdd(scheduledScaler)
	}
}

// scheduledScalerHpaCronAdd will update the hpa when the scheduled scaler fires.
func (c *ScheduledScalerController) scheduledScalerHpaCronAdd(scheduledScaler *scalingv1alpha1.ScheduledScaler) {
	tz := scheduledScaler.Spec.TimeZone
	ss, err := c.scheduledScalersLister.ScheduledScalers(scheduledScaler.Namespace).Get(scheduledScaler.Name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf(
			"FAILED TO GET SCHEDULED SCALER: %s - %w", scheduledScaler.Spec.Target.Name, err))
		return
	}
	hpaClient := c.kubeClient.AutoscalingV1().HorizontalPodAutoscalers(scheduledScaler.Namespace)
	hpa, err := hpaClient.Get(scheduledScaler.Spec.Target.Name, metav1.GetOptions{})
	if err != nil {
		utilruntime.HandleError(fmt.Errorf(
			"FAILED TO GET HPA: %s - %w", scheduledScaler.Spec.Target.Name, err))
		return
	}

	ssClient := c.restdevClient.ScalingV1alpha1().ScheduledScalers(scheduledScaler.Namespace)
	// TODO: is this really needed?
	ssCopy := ss.DeepCopy()
	stepsCron := c.cronProxy.Create(tz)
	var mutex sync.Mutex
	for key := range ssCopy.Spec.Steps {
		step := scheduledScaler.Spec.Steps[key]
		min, max := scalingstep.Parse(step)
		c.cronProxy.Push(stepsCron, step.Runat, func() {
			// If this scheduled scaler retries, don't let the "next" one get overwritten by its retry.
			mutex.Lock()
			defer mutex.Unlock()

			hpaRetries := 0
		HpaAgain:
			if hpaRetries > c.maxRetries {
				utilruntime.HandleError(fmt.Errorf(
					"FAILED TO UPDATE HPA: %s after %d retries", scheduledScaler.Spec.Target.Name, hpaRetries))
				return
			}
			hpa, err = hpaClient.Get(scheduledScaler.Spec.Target.Name, metav1.GetOptions{})
			if err != nil {
				utilruntime.HandleError(fmt.Errorf(
					"FAILED TO UPDATE HPA: %s - %w", scheduledScaler.Spec.Target.Name, err))
				return
			}
			hpa.Spec.MinReplicas = min
			hpa.Spec.MaxReplicas = *max
			_, err = hpaClient.Update(hpa)
			if apierr.IsConflict(err) {
				glog.Infof("FAILED TO UPDATE HPA: %s - %w; retrying", scheduledScaler.Spec.Target.Name, err)
				hpaRetries++
				goto HpaAgain
			}
			if err != nil {
				utilruntime.HandleError(fmt.Errorf(
					"FAILED TO UPDATE HPA: %s - %w", scheduledScaler.Spec.Target.Name, err))
				return
			}
			ssRetries := 0
		SSAgain:
			if ssRetries > c.maxRetries {
				utilruntime.HandleError(fmt.Errorf(
					"FAILED TO UPDATE SS: %s after %d retries", scheduledScaler.Name, ssRetries))
				return
			}
			ss, err := ssClient.Get(scheduledScaler.Name, metav1.GetOptions{})
			if err != nil {
				utilruntime.HandleError(fmt.Errorf(
					"FAILED TO UPDATE SCHEDULED SCALER STATUS: %s - %w", scheduledScaler.Name, err))
				return
			}
			ss.Status.Mode = step.Mode
			ss.Status.MinReplicas = *min
			ss.Status.MaxReplicas = *max
			_, err = ssClient.Update(ss)
			if apierr.IsConflict(err) {
				glog.Infof("FAILED TO UPDATE SCHEDULED SCALER STATUS: %s - %w; retrying", scheduledScaler.Name, err)
				ssRetries++
				goto SSAgain
			}
			if err != nil {
				utilruntime.HandleError(fmt.Errorf(
					"FAILED TO UPDATE SCHEDULED SCALER STATUS: %s - %w", scheduledScaler.Name, err))
				return
			}
			glog.Infof("SETTING RANGE SCALER: %s/%s -> %s - %d:%d", scheduledScaler.Namespace, scheduledScaler.Name, scheduledScaler.Spec.Target.Name, *min, *max)
		})
	}
	c.cronProxy.Start(stepsCron)
	c.scheduledScalerTargets = append(c.scheduledScalerTargets, ScheduledScalerTarget{scheduledScaler.Spec.Target.Name, scheduledScaler.Spec.Target.Kind, stepsCron})
	glog.Infof("SCHEDULED SCALER CREATED: %s -> %s", scheduledScaler.Name, scheduledScaler.Spec.Target.Name)
}

func (c *ScheduledScalerController) scheduledScalerIgCronAdd(scheduledScaler *scalingv1alpha1.ScheduledScaler) {
	projectId, zone, _ := scalingmetadata.GetClusterInfo()
	tz := scheduledScaler.Spec.TimeZone
	ss, err := c.scheduledScalersLister.ScheduledScalers(scheduledScaler.Namespace).Get(scheduledScaler.Name)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	ctx := context.Background()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	computeService, err := compute.New(client)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	autoscaler, err := computeService.Autoscalers.Get(projectId, zone, scheduledScaler.Spec.Target.Name).Do()
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	ssCopy := ss.DeepCopy()
	stepsCron := c.cronProxy.Create(tz)
	for key := range scheduledScaler.Spec.Steps {
		step := scheduledScaler.Spec.Steps[key]
		min, max := scalingstep.Parse(step)
		c.cronProxy.Push(stepsCron, step.Runat, func() {
			autoscaler, err = computeService.Autoscalers.Get(projectId, zone, scheduledScaler.Spec.Target.Name).Do()
			autoscaler.AutoscalingPolicy.MaxNumReplicas = int64(*max)
			autoscaler.AutoscalingPolicy.MinNumReplicas = int64(*min)
			_, err := computeService.Autoscalers.Update(projectId, zone, autoscaler).Do()
			if err != nil {
				utilruntime.HandleError(fmt.Errorf(
					"FAILED TO UPDATE IG AUTOSCALER: %s - %w", scheduledScaler.Spec.Target.Name, err))
				return
			}
			ssCopy.Status.Mode = step.Mode
			ssCopy.Status.MinReplicas = *min
			ssCopy.Status.MaxReplicas = *max
			if _, err = c.restdevClient.ScalingV1alpha1().ScheduledScalers(scheduledScaler.Namespace).Update(ssCopy); err != nil {
				utilruntime.HandleError(fmt.Errorf(
					"FAILED TO UPDATE SCHEDULED SCALER STATUS: %s - %w", scheduledScaler.Name, err))
			}
			glog.Infof("SETTING RANGE IG SCALER: %s -> %s - %d/%d", scheduledScaler.Name, scheduledScaler.Spec.Target.Name, *min, *max)
		})
	}

	c.cronProxy.Start(stepsCron)
	c.scheduledScalerTargets = append(c.scheduledScalerTargets, ScheduledScalerTarget{scheduledScaler.Spec.Target.Name, scheduledScaler.Spec.Target.Kind, stepsCron})
	glog.Infof("SCHEDULED SCALER CREATED: %s -> %s", scheduledScaler.Name, scheduledScaler.Spec.Target.Name)
}

/**
 *
 * Update methods
 *
 * These methods will handle updates to existing resources
 *
 */
func (c *ScheduledScalerController) scheduledScalerUpdate(old, new interface{}) {
	oldScheduledScaler := old.(*scalingv1alpha1.ScheduledScaler)
	newScheduledScaler := new.(*scalingv1alpha1.ScheduledScaler)
	if reflect.DeepEqual(oldScheduledScaler.Spec, newScheduledScaler.Spec) {
		return
	}
	c.scheduledScalerDelete(old)
	c.scheduledScalerAdd(new)
}

/**
 *
 * Delete methods
 *
 * These methods will handle deletion of resources
 *
 */
func (c *ScheduledScalerController) scheduledScalerDelete(obj interface{}) {
	c.scheduledScalerCronDelete(obj)
}

func (c *ScheduledScalerController) scheduledScalerCronDelete(obj interface{}) {
	scheduledScaler := obj.(*scalingv1alpha1.ScheduledScaler)
	// find index
	key, err := c.scheduledScalerFindTargetKey(scheduledScaler.Spec.Target.Name)
	if err {
		glog.Infof("FAILED TO DELETE SCALER TARGET: %s -> %s (NotFound)", scheduledScaler.Name, scheduledScaler.Spec.Target.Name)
		return
	}
	glog.Infof("STOPPING CRONS FOR SCALER TARGET: %s -> %s", scheduledScaler.Name, scheduledScaler.Spec.Target.Name)
	c.cronProxy.Stop(c.scheduledScalerTargets[key].Cron)
	c.scheduledScalerTargets[key] = c.scheduledScalerTargets[0]
	c.scheduledScalerTargets = c.scheduledScalerTargets[1:]
	glog.Infof("SCHEDULED SCALER TARGET DELETED: %s -> %s", scheduledScaler.Name, scheduledScaler.Spec.Target.Name)
}

func (c *ScheduledScalerController) scheduledScalerFindTargetKey(name string) (int, bool) {
	for key := range c.scheduledScalerTargets {
		if c.scheduledScalerTargets[key].Name == name {
			return key, false
		}
	}

	return -1, true
}

/**
 *
 * Create new instance of the Scaling Controller
 *
 */
func NewScheduledScalerController(
	maxRetries int,
	cronProxy scalingcron.CronProxy,
	informerFactory informers.SharedInformerFactory,
	restdevClient clientset.Interface,
	kubeClient kubernetes.Interface,
) *ScheduledScalerController {
	scheduledScalerInformer := informerFactory.Scaling().V1alpha1().ScheduledScalers()
	scheduledScalersLister := scheduledScalerInformer.Lister()
	var scheduledScalerTargets []ScheduledScalerTarget

	c := &ScheduledScalerController{
		maxRetries:             maxRetries,
		cronProxy:              cronProxy,
		informerFactory:        informerFactory,
		restdevClient:          restdevClient,
		kubeClient:             kubeClient,
		scheduledScalersLister: scheduledScalersLister,
		scheduledScalersSynced: scheduledScalerInformer.Informer().HasSynced,
		scheduledScalerTargets: scheduledScalerTargets,
	}
	scheduledScalerInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.scheduledScalerAdd,
			UpdateFunc: c.scheduledScalerUpdate,
			DeleteFunc: c.scheduledScalerDelete,
		},
	)
	return c
}

/**
 *
 * Run the app
 *
 */
func main() {
	var (
		kubeconfig string
		maxRetries int
	)

	flag.Set("logtostderr", "true")
	flag.IntVar(&maxRetries, "max-retries", 1, "maximum retries before failing to update HPA or SS")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	//clientset, err := kubernetes.NewForConfig(config)
	kubeClient, err := kubernetes.NewForConfig(config)
	restdevClient, err := clientset.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}

	factory := informers.NewSharedInformerFactory(restdevClient, time.Hour*24)
	cronProxy := new(scalingcron.CronImpl)
	controller := NewScheduledScalerController(maxRetries, cronProxy, factory, restdevClient, kubeClient)
	stop := make(chan struct{})
	defer close(stop)
	err = controller.Run(stop)
	if err != nil {
		glog.Fatal(err)
	}
	select {}
}
