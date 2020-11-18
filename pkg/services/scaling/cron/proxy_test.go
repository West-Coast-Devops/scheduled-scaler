package cron

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCronImpl_Create(t *testing.T) {
	tests := []struct {
		name         string
		timeZone     string
		wantLocation *time.Location
		wantErr      bool
	}{
		{
			name:         `empty is UTC`,
			wantLocation: time.UTC,
		},
		{
			name:         `Local is Local`,
			timeZone:     "Local",
			wantLocation: time.Local,
		},
		{
			name:     `bogus timezone returns error`,
			timeZone: "bogus",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := &CronImpl{}
			got, err := ci.Create(tt.timeZone)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantLocation, got.Location())
		})
	}
}
