package cron

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCronParse(t *testing.T) {
	cases := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{
			name:    `empty`,
			wantErr: true,
		},
		{
			name: `Includes seconds`,
			spec: "0 0 19 * 11 TUE",
		},
		{
			name: `Includes seconds`,
			spec: "0 0 19 * 11 TUE",
		},
		{
			name: `with CRON_TZ`,
			spec: "CRON_TZ=America/Los_Angeles 0 0 19 * 11 TUE",
		},
		{
			name: `with CRON_TZ`,
			spec: "TZ=America/Los_Angeles 0 0 19 * 11 TUE",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := V1Parser.Parse(tt.spec)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, schedule)
		})
	}
}
