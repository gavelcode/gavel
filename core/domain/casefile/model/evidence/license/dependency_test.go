package license_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/usegavel/gavel/core/domain/casefile/model/evidence/license"
)

func TestNewDependencyLicense(t *testing.T) {
	tests := []struct {
		name    string
		depName string
		version string
		license string
		wantErr bool
	}{
		{
			name:    "shouldCreateValidDependencyLicense",
			depName: "github.com/stretchr/testify",
			version: "viol1.9.0",
			license: "MIT",
		},
		{
			name:    "shouldRejectEmptyName",
			depName: "",
			version: "viol1.0.0",
			license: "MIT",
			wantErr: true,
		},
		{
			name:    "shouldRejectBlankName",
			depName: "   ",
			version: "viol1.0.0",
			license: "MIT",
			wantErr: true,
		},
		{
			name:    "shouldRejectEmptyVersion",
			depName: "some-lib",
			version: "",
			license: "MIT",
			wantErr: true,
		},
		{
			name:    "shouldRejectBlankVersion",
			depName: "some-lib",
			version: "  ",
			license: "MIT",
			wantErr: true,
		},
		{
			name:    "shouldRejectEmptyLicense",
			depName: "some-lib",
			version: "viol1.0.0",
			license: "",
			wantErr: true,
		},
		{
			name:    "shouldRejectBlankLicense",
			depName: "some-lib",
			version: "viol1.0.0",
			license: "   ",
			wantErr: true,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			depLic, err := license.NewDependency(tcase.depName, tcase.version, tcase.license)

			if tcase.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, license.ErrInvalidDependency)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tcase.depName, depLic.Name())
			assert.Equal(t, tcase.version, depLic.Version())
			assert.Equal(t, tcase.license, depLic.License())
		})
	}
}
