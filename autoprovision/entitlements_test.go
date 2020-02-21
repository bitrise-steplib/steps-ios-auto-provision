package autoprovision_test

import (
	"testing"

	"github.com/bitrise-steplib/steps-ios-auto-provision/autoprovision"
	"github.com/stretchr/testify/require"
)

func TestICloudContainers(t *testing.T) {
	tests := []struct {
		name                string
		projectEntitlements autoprovision.Entitlement
		want                []string
		errHandler          func(require.TestingT, error, ...interface{})
	}{
		{
			name:                "no containers",
			projectEntitlements: autoprovision.Entitlement(map[string]interface{}{}),
			want:                nil,
			errHandler:          require.NoError,
		},
		{
			name: "no containers - CloudDocuments",
			projectEntitlements: autoprovision.Entitlement(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudDocuments",
				},
			}),
			want:       nil,
			errHandler: require.NoError,
		},
		{
			name: "no containers - CloudKit",
			projectEntitlements: autoprovision.Entitlement(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudKit",
				},
			}),
			want:       nil,
			errHandler: require.NoError,
		},
		{
			name: "no containers - CloudKit and CloudDocuments",
			projectEntitlements: autoprovision.Entitlement(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudKit",
					"CloudDocuments",
				},
			}),
			want:       nil,
			errHandler: require.NoError,
		},
		{
			name: "has containers - CloudDocuments",
			projectEntitlements: autoprovision.Entitlement(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudDocuments",
				},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.test.container.id",
					"iCloud.test.container.id2"},
			}),
			want:       []string{"iCloud.test.container.id", "iCloud.test.container.id2"},
			errHandler: require.NoError,
		},
		{
			name: "has containers - CloudKit",
			projectEntitlements: autoprovision.Entitlement(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudKit",
				},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.test.container.id",
					"iCloud.test.container.id2"},
			}),
			want:       []string{"iCloud.test.container.id", "iCloud.test.container.id2"},
			errHandler: require.NoError,
		},
		{
			name: "has containers - CloudKit and CloudDocuments",
			projectEntitlements: autoprovision.Entitlement(map[string]interface{}{
				"com.apple.developer.icloud-services": []interface{}{
					"CloudKit",
					"CloudDocuments",
				},
				"com.apple.developer.icloud-container-identifiers": []interface{}{
					"iCloud.test.container.id",
					"iCloud.test.container.id2"},
			}),
			want:       []string{"iCloud.test.container.id", "iCloud.test.container.id2"},
			errHandler: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.projectEntitlements.ICloudContainers()
			require.Equal(t, got, tt.want)
			tt.errHandler(t, err)
		})
	}
}
