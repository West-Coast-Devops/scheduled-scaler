package main

import (
	"flag"
	"fmt"
	"time"
	"reflect"
	"strings"
	"io/ioutil"

	"net/http"

	"github.com/golang/glog"
	"github.com/robfig/cron"

	//kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	
	clientset "k8s.restdev.com/operators/pkg/client/clientset/versioned"
	informers "k8s.restdev.com/operators/pkg/client/informers/externalversions"
	listers "k8s.restdev.com/operators/pkg/client/listers/scaling/v1alpha1"
	scalingv1alpha1 "k8s.restdev.com/operators/pkg/apis/scaling/v1alpha1"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

/**
 *
 * Target Struct
 *
 */
type ScheduledScalerTarget struct {
	Name string
	Type string
	Cron *cron.Cron
}

/**
 *
 * Scaling Controller Struct
 *
 */
type ScheduledScalerController struct {
	informerFactory informers.SharedInformerFactory
	restdevClient clientset.Interface
	kubeClient *kubernetes.Clientset
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

/**
 *
 * Add methods
 *
 * These methods will handle new resources
 *
 */
func (c *ScheduledScalerController) scheduledScalerAdd(obj interface{}) {
				scheduledScaler := obj.(*scalingv1alpha1.ScheduledScaler)
				if scheduledScaler.Spec.Target.Type == "hpa" {
					c.scheduledScalerHpaCronAdd(obj)
				} else if scheduledScaler.Spec.Target.Type == "ig" {
					c.scheduledScalerIgCronAdd(obj)
				}
}

func (c *ScheduledScalerController) scheduledScalerHpaCronAdd(obj interface{}) {
	scheduledScaler := obj.(*scalingv1alpha1.ScheduledScaler)
	tz := scheduledScaler.Spec.TimeZone
	ss, err := c.scheduledScalersLister.ScheduledScalers(scheduledScaler.Namespace).Get(scheduledScaler.Name)
	if err != nil {
		panic(err.Error())
	}
	hpaClient := c.kubeClient.AutoscalingV1().HorizontalPodAutoscalers(scheduledScaler.Namespace)
	hpa, err := hpaClient.Get(scheduledScaler.Spec.Target.Name, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	ssCopy := ss.DeepCopy()
	l, _ := time.LoadLocation(tz)
	stepsCron := cron.NewWithLocation(l)
	for key := range ssCopy.Spec.Steps {
		step := scheduledScaler.Spec.Steps[key]
		s, _ := cron.Parse(step.Runat)
		if step.Mode == "range" {
			stepsCron.Schedule(s, cron.FuncJob(func() {
				hpa, err = hpaClient.Get(scheduledScaler.Spec.Target.Name, metav1.GetOptions{})
				if err != nil {
					panic(err.Error())
				}
				hpa.Spec.MinReplicas = step.MinReplicas
				hpa.Spec.MaxReplicas = *step.MaxReplicas
				_, err := hpaClient.Update(hpa)
				if err != nil {
					glog.Infof("FAILED TO UPDATE HPA: %s - %s", scheduledScaler.Spec.Target.Name, err.Error())
				} else {
					ssCopy.Status.Mode = step.Mode
					ssCopy.Status.MinReplicas = *step.MinReplicas
					ssCopy.Status.MaxReplicas = *step.MaxReplicas
					_, err := c.restdevClient.ScalingV1alpha1().ScheduledScalers(scheduledScaler.Namespace).Update(ssCopy)
					if err != nil {
						glog.Infof("FAILED TO UPDATE SCHEDULED SCALER STATUS: %s - %s", scheduledScaler.Name, err.Error())
					}
					glog.Infof("SETTING RANGE SCALER: %s/%s -> %s - %d:%d", scheduledScaler.Namespace, scheduledScaler.Name, scheduledScaler.Spec.Target.Name, *step.MinReplicas, *step.MaxReplicas)
				}
			}))
		}
		if step.Mode == "fixed" {
			stepsCron.Schedule(s, cron.FuncJob(func() {
				hpa, err = hpaClient.Get(scheduledScaler.Spec.Target.Name, metav1.GetOptions{})
				if err != nil {
					panic(err.Error())
				}
				hpa.Spec.MinReplicas = step.Replicas
        hpa.Spec.MaxReplicas = *step.Replicas
        _, err := hpaClient.Update(hpa)
        if err != nil {
          glog.Infof("FAILED TO UPDATE HPA: %s - %s", scheduledScaler.Spec.Target.Name, err.Error())
        } else {
					ssCopy.Status.Mode = step.Mode
					ssCopy.Status.MinReplicas = *step.Replicas
					ssCopy.Status.MaxReplicas = *step.Replicas
					_, err := c.restdevClient.ScalingV1alpha1().ScheduledScalers(scheduledScaler.Namespace).Update(ssCopy)
					if err != nil {
						glog.Infof("FAILED TO UPDATE SCHEDULED SCALER STATUS: %s - %s", scheduledScaler.Name, err.Error())
					}
					glog.Infof("SETTING FIXED SCALER: %s/%s -> %s - %d", scheduledScaler.Namespace, scheduledScaler.Name, scheduledScaler.Spec.Target.Name, *step.Replicas)
				}
			}))
		}
	}
	stepsCron.Start()
	c.scheduledScalerTargets = append(c.scheduledScalerTargets, ScheduledScalerTarget{scheduledScaler.Spec.Target.Name, scheduledScaler.Spec.Target.Type, stepsCron})
	glog.Infof("SCHEDULED SCALER CREATED: %s -> %s", scheduledScaler.Name, scheduledScaler.Spec.Target.Name)
}

func (c *ScheduledScalerController) scheduledScalerIgCronAdd(obj interface{}) {
	httpclient := &http.Client{}
	projectIdReq, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
	projectIdReq.Header.Add("Metadata-Flavor", "Google")
	projectIdResp, err := httpclient.Do(projectIdReq)
	zoneReq, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/zone", nil)
	zoneReq.Header.Add("Metadata-Flavor", "Google")
	zoneResp, err := httpclient.Do(zoneReq)
	defer zoneResp.Body.Close()
	defer projectIdResp.Body.Close()
	projectIdBody, err := ioutil.ReadAll(projectIdResp.Body)
	projectId := string(projectIdBody)
	zoneBody, err := ioutil.ReadAll(zoneResp.Body)
	zoneSlice := strings.Split(string(zoneBody), "/")
	zone := zoneSlice[ len(zoneSlice) - 1 ]

	scheduledScaler := obj.(*scalingv1alpha1.ScheduledScaler)
	tz := scheduledScaler.Spec.TimeZone
	ss, err := c.scheduledScalersLister.ScheduledScalers(scheduledScaler.Namespace).Get(scheduledScaler.Name)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		panic(err.Error())
	}
	computeService, err := compute.New(client)
	if err != nil {
		panic(err.Error())
	}
	
	autoscaler, err := computeService.Autoscalers.Get(projectId, zone, scheduledScaler.Spec.Target.Name).Do()
	if err != nil {
		panic(err.Error())
	}

	ssCopy := ss.DeepCopy()
	l, _ := time.LoadLocation(tz)
	stepsCron := cron.NewWithLocation(l)
	for key := range scheduledScaler.Spec.Steps {
		step := scheduledScaler.Spec.Steps[key]
		s, _ := cron.Parse(step.Runat)
		if step.Mode == "range" {
			stepsCron.Schedule(s, cron.FuncJob(func() {
				autoscaler, err = computeService.Autoscalers.Get(projectId, zone, scheduledScaler.Spec.Target.Name).Do()
				autoscaler.AutoscalingPolicy.MaxNumReplicas = int64(*step.MaxReplicas)
				autoscaler.AutoscalingPolicy.MinNumReplicas = int64(*step.MinReplicas)
				_, err := computeService.Autoscalers.Update(projectId, zone, autoscaler).Do()
				if err != nil {
					glog.Infof("FAILED TO UPDATE IG AUTOSCALER: %s - %s", scheduledScaler.Spec.Target.Name, err.Error())
				} else {
					ssCopy.Status.Mode = step.Mode
					ssCopy.Status.MinReplicas = *step.MinReplicas
					ssCopy.Status.MaxReplicas = *step.MaxReplicas
					_, err := c.restdevClient.ScalingV1alpha1().ScheduledScalers(scheduledScaler.Namespace).Update(ssCopy)
					if err != nil {
						glog.Infof("FAILED TO UPDATE SCHEDULED SCALER STATUS: %s - %s", scheduledScaler.Name, err.Error())
					}
					glog.Infof("SETTING RANGE IG SCALER: %s -> %s - %d/%d", scheduledScaler.Name, scheduledScaler.Spec.Target.Name, *step.MinReplicas, *step.MaxReplicas)
				}
			}))
		}
		if step.Mode == "fixed" {
			stepsCron.Schedule(s, cron.FuncJob(func() {
				autoscaler, err = computeService.Autoscalers.Get(projectId, zone, scheduledScaler.Spec.Target.Name).Do()
				autoscaler.AutoscalingPolicy.MaxNumReplicas = int64(*step.Replicas)
				autoscaler.AutoscalingPolicy.MinNumReplicas = int64(*step.Replicas)
				_, err := computeService.Autoscalers.Update(projectId, zone, autoscaler).Do()
				if err != nil {
					glog.Infof("FAILED TO UPDATE IG AUTOSCALER: %s - %s", scheduledScaler.Spec.Target.Name, err.Error())
				} else {
					ssCopy.Status.Mode = step.Mode
					ssCopy.Status.MinReplicas = *step.Replicas
					ssCopy.Status.MaxReplicas = *step.Replicas
					_, err := c.restdevClient.ScalingV1alpha1().ScheduledScalers(scheduledScaler.Namespace).Update(ssCopy)
					if err != nil {
						glog.Infof("FAILED TO UPDATE SCHEDULED SCALER STATUS: %s - %s", scheduledScaler.Name, err.Error())
					}
					glog.Infof("SETTING RANGE IG SCALER: %s -> %s - %d", scheduledScaler.Name, scheduledScaler.Spec.Target.Name, *step.Replicas)
				}
			}))
		}
	}

	stepsCron.Start()
	c.scheduledScalerTargets = append(c.scheduledScalerTargets, ScheduledScalerTarget{scheduledScaler.Spec.Target.Name, scheduledScaler.Spec.Target.Type, stepsCron})
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
	c.scheduledScalerTargets[key].Cron.Stop()
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
	informerFactory informers.SharedInformerFactory,
	restdevClient clientset.Interface,
	kubeClient *kubernetes.Clientset,
) *ScheduledScalerController {
	scheduledScalerInformer := informerFactory.Scaling().V1alpha1().ScheduledScalers()
	scheduledScalersLister := scheduledScalerInformer.Lister()
	var scheduledScalerTargets []ScheduledScalerTarget

	c := &ScheduledScalerController{
		informerFactory: informerFactory,
		restdevClient: restdevClient,
		kubeClient: kubeClient,
		scheduledScalersLister: scheduledScalersLister,
		scheduledScalersSynced: scheduledScalerInformer.Informer().HasSynced,
		scheduledScalerTargets: scheduledScalerTargets,
	}
	scheduledScalerInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.scheduledScalerAdd,
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
	var kubeconfig string

	flag.Set("logtostderr", "true")
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
	controller := NewScheduledScalerController(factory, restdevClient, kubeClient)
	stop := make(chan struct{})
	defer close(stop)
	err = controller.Run(stop)
	if err != nil {
		glog.Fatal(err)
	}
	select {}
}
