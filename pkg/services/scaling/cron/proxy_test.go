package cron

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCronImpl_Create(t *testing.T) {

	// Set up test expectation objects in outer testing scope.
	losAngelesTZ, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)
	assert.Equal(t, "America/Los_Angeles", losAngelesTZ.String())
	newYorkTZ, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	assert.NotEqual(t, losAngelesTZ, newYorkTZ)
	assert.Equal(t, "America/New_York", newYorkTZ.String())

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
		{
			name:         `PST`,
			timeZone:     "America/Los_Angeles",
			wantLocation: losAngelesTZ,
		},
		{
			name:         `EST`,
			timeZone:     "America/New_York",
			wantLocation: newYorkTZ,
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
