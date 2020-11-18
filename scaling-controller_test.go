package main

import (
	"github.com/golang/mock/gomock"
	"github.com/robfig/cron"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	scalingv1alpha1 "k8s.restdev.com/operators/pkg/apis/scaling/v1alpha1"
	"k8s.restdev.com/operators/pkg/client/clientset/versioned/mock_versioned"
	mock_v1alpha12 "k8s.restdev.com/operators/pkg/client/clientset/versioned/typed/scaling/v1alpha1/mock_v1alpha1"
	"k8s.restdev.com/operators/pkg/client/listers/scaling/v1alpha1/mock_v1alpha1"
	cron2 "k8s.restdev.com/operators/pkg/services/scaling/cron"
	"k8s.restdev.com/operators/pkg/services/scaling/cron/mock_cron"
	mock_external "k8s.restdev.com/operators/test/external"
	"testing"
)

func Test_validateScheduledScaler(t *testing.T) {
	tests := []struct {
		name            string
		scheduledScaler *scalingv1alpha1.ScheduledScaler
		wantErr         bool
	}{
		{
			name:    `nil scheduledScaler`,
			wantErr: true,
		},
		{
			name: `Runat with too few fields`,
			scheduledScaler: &scalingv1alpha1.ScheduledScaler{
				Spec: scalingv1alpha1.ScheduledScalerSpec{
					Steps: []scalingv1alpha1.ScheduledScalerStep{
						{
							Runat: "1 2 3",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: `Runat with February 30`,
			scheduledScaler: &scalingv1alpha1.ScheduledScaler{
				Spec: scalingv1alpha1.ScheduledScalerSpec{
					Steps: []scalingv1alpha1.ScheduledScalerStep{
						{
							Runat: "0 0 6 30 FEB *",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: `valid date is ok ("0 0 6 * * SAT")`,
			scheduledScaler: &scalingv1alpha1.ScheduledScaler{
				Spec: scalingv1alpha1.ScheduledScalerSpec{
					Steps: []scalingv1alpha1.ScheduledScalerStep{
						{
							Runat: "0 0 6 * * SAT",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ScheduledScalerController{
				cronProxy: new(cron2.CronImpl),
			}
			err := c.validateScheduledScaler(tt.scheduledScaler)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func newInt32(i int32) *int32 {
	return &i
}

func TestScheduledScalerController_scheduledScalerHpaCronAdd(t *testing.T) {
	testHPA := &v1.HorizontalPodAutoscaler{
		Spec: v1.HorizontalPodAutoscalerSpec{
			MinReplicas: newInt32(10),
			MaxReplicas: 20,
		},
	}
	testSS := scalingv1alpha1.ScheduledScaler{
		Spec: scalingv1alpha1.ScheduledScalerSpec{
			Steps: []scalingv1alpha1.ScheduledScalerStep{
				{
					Mode:        "range",
					Runat:       "0 0 6 * * SAT",
					MinReplicas: newInt32(100),
					MaxReplicas: newInt32(200),
				},
			},
		},
	}
	tests := []struct {
		name           string
		ss             scalingv1alpha1.ScheduledScaler
		ssListerGetErr error
		hpaGetResults  []*v1.HorizontalPodAutoscaler
		hpaGetErrs     []error
		hpaUpdatesErrs []error
		ssGetResults   []*scalingv1alpha1.ScheduledScaler
		ssGetErrs      []error
		ssUpdateErrs   []error
	}{
		{
			name: "empty SS",
			hpaGetResults: []*v1.HorizontalPodAutoscaler{
				nil,
			},
			hpaGetErrs: []error{
				nil,
			},
		},
		{
			name: "SS with one spec, fails once and retries",
			ss:   testSS,
			hpaGetResults: []*v1.HorizontalPodAutoscaler{
				testHPA,
				testHPA,
				testHPA,
			},
			hpaGetErrs: []error{
				nil,
				nil,
				nil,
			},
			hpaUpdatesErrs: []error{
				errors.NewConflict(schema.GroupResource{}, "foo", nil),
				nil,
			},
			ssUpdateErrs: []error{
				nil,
			},
			ssGetResults: []*scalingv1alpha1.ScheduledScaler{
				&testSS,
			},
			ssGetErrs: []error{
				nil,
			},
		},
		{
			name: "SS with one spec, fails ss update once and retries",
			ss:   testSS,
			hpaGetResults: []*v1.HorizontalPodAutoscaler{
				testHPA,
				testHPA,
			},
			hpaGetErrs: []error{
				nil,
				nil,
			},
			hpaUpdatesErrs: []error{
				nil,
			},
			ssUpdateErrs: []error{
				errors.NewConflict(schema.GroupResource{}, "foo", nil),
				nil,
			},
			ssGetResults: []*scalingv1alpha1.ScheduledScaler{
				&testSS,
				&testSS,
			},
			ssGetErrs: []error{
				nil,
				nil,
			},
		},
		{
			name: "SS with one spec, fails more than maxRetries; never calls ss update",
			ss:   testSS,
			hpaGetResults: []*v1.HorizontalPodAutoscaler{
				testHPA,
				testHPA,
				testHPA,
				testHPA,
				testHPA,
			},
			hpaGetErrs: []error{
				nil,
				nil,
				nil,
				nil,
				nil,
			},
			hpaUpdatesErrs: []error{
				errors.NewConflict(schema.GroupResource{}, "foo", nil),
				errors.NewConflict(schema.GroupResource{}, "foo", nil),
				errors.NewConflict(schema.GroupResource{}, "foo", nil),
				errors.NewConflict(schema.GroupResource{}, "foo", nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCronProxy := mock_cron.NewMockCronProxy(ctrl)
			mockCronProxy.EXPECT().
				Create(gomock.Any()).
				DoAndReturn(func(tz string) *cron.Cron {
					return cron.New()
				})
			var calls []func()
			for _, step := range tt.ss.Spec.Steps {
				mockCronProxy.EXPECT().
					Push(gomock.Any(), gomock.Eq(step.Runat), gomock.Any()).
					Do(func(c *cron.Cron, time string, call func()) {
						calls = append(calls, call)
					})
			}
			mockCronProxy.EXPECT().
				Start(gomock.Any()).
				Do(func(c *cron.Cron) {
					for _, call := range calls {
						call()
					}
				})

			mockNsLister := mock_v1alpha1.NewMockScheduledScalerNamespaceLister(ctrl)
			mockNsLister.EXPECT().
				Get(tt.ss.Name).
				Return(&tt.ss, tt.ssListerGetErr)
			mockLister := mock_v1alpha1.NewMockScheduledScalerLister(ctrl)
			mockLister.EXPECT().
				ScheduledScalers(tt.ss.Namespace).
				Return(mockNsLister)

			mockHPA := mock_external.NewMockHorizontalPodAutoscalerInterface(ctrl)
			getCallIndex := 0
			mockHPA.EXPECT().
				Get(gomock.Any(), gomock.Any()).
				Times(len(tt.hpaGetResults)).
				DoAndReturn(func(name string, options metav1.GetOptions) (*v1.HorizontalPodAutoscaler, error) {
					result := tt.hpaGetResults[getCallIndex]
					err := tt.hpaGetErrs[getCallIndex]
					getCallIndex++
					return result, err
				})
			updateCallIndex := 0
			mockHPA.EXPECT().
				Update(gomock.Any()).
				Times(len(tt.hpaUpdatesErrs)).
				DoAndReturn(func(input *v1.HorizontalPodAutoscaler) (*v1.HorizontalPodAutoscaler, error) {
					err := tt.hpaUpdatesErrs[updateCallIndex]
					updateCallIndex++
					return input, err
				})
			mockAutoscaling := mock_external.NewMockAutoscalingV1Interface(ctrl)
			mockAutoscaling.EXPECT().
				HorizontalPodAutoscalers(tt.ss.Namespace).
				Return(mockHPA)
			mockKubeClient := mock_external.NewMockInterface(ctrl)
			mockKubeClient.EXPECT().
				AutoscalingV1().
				Return(mockAutoscaling)

			mockScheduledScalerInterface := mock_v1alpha12.NewMockScheduledScalerInterface(ctrl)
			ssGetIndex := 0
			mockScheduledScalerInterface.EXPECT().
				Get(gomock.Any(), gomock.Any()).
				Times(len(tt.ssGetResults)).
				DoAndReturn(func(name string, options metav1.GetOptions) (*scalingv1alpha1.ScheduledScaler, error) {
					result := tt.ssGetResults[ssGetIndex]
					err := tt.ssGetErrs[ssGetIndex]
					ssGetIndex++
					return result, err
				})

			ssUpdateIndex := 0
			mockScheduledScalerInterface.EXPECT().
				Update(gomock.Any()).
				Times(len(tt.ssUpdateErrs)).
				DoAndReturn(func(*scalingv1alpha1.ScheduledScaler) (*scalingv1alpha1.ScheduledScaler, error) {
					err := tt.ssUpdateErrs[ssUpdateIndex]
					ssUpdateIndex++
					return nil, err
				})
			mockScalingV1alpha1Interface := mock_v1alpha12.NewMockScalingV1alpha1Interface(ctrl)
			mockScalingV1alpha1Interface.EXPECT().
				ScheduledScalers(tt.ss.Namespace).
				Return(mockScheduledScalerInterface)
			mockRestdevClient := mock_versioned.NewMockInterface(ctrl)
			mockRestdevClient.EXPECT().
				ScalingV1alpha1().
				Return(mockScalingV1alpha1Interface)
			c := &ScheduledScalerController{
				cronProxy:              mockCronProxy,
				restdevClient:          mockRestdevClient,
				kubeClient:             mockKubeClient,
				scheduledScalersLister: mockLister,
			}
			c.scheduledScalerHpaCronAdd(&tt.ss)
		})
	}
}
