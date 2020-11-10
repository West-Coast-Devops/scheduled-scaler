package main

import (
	"github.com/stretchr/testify/require"
	scalingv1alpha1 "k8s.restdev.com/operators/pkg/apis/scaling/v1alpha1"
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
			err := validateScheduledScaler(tt.scheduledScaler)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
