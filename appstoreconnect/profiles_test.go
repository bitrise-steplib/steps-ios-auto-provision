package appstoreconnect

import "testing"

func TestProfile_isXcodeManaged(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		want    bool
	}{
		{
			name: "Non Xcode Managed - Bitrise Develpment (io.bitrise.sample)",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Bitrise Develpment (io.bitrise.sample)",
				},
			},
			want: false,
		},
		{
			name: "Non Xcode Managed - Bitrise App Store (io.bitrise.sample)",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Bitrise App Store (io.bitrise.sample)",
				},
			},
			want: false,
		},
		{
			name: "Non Xcode Managed - Bitrise Ad Hoc (io.bitrise.sample)",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Bitrise Ad Hoc (io.bitrise.sample)",
				},
			},
			want: false,
		},
		{
			name: "Xcode Managed - XC Ad Hoc: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "XC Ad Hoc: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - XC: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "XC: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - XC Ad Hoc: { bundle id }",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "XC Ad Hoc: { bundle id }",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - XC: { bundle id }",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "XC: { bundle id }",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - iOS Team Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "iOS Team Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - iOS Team Ad Hoc Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "iOS Team Ad Hoc Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - iOS Team Ad Hoc Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "iOS Team Ad Hoc Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - iOS Team Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "iOS Team Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - tvOS Team Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "tvOS Team Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - tvOS Team Ad Hoc Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "tvOS Team Ad Hoc Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - tvOS Team Ad Hoc Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "tvOS Team Ad Hoc Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - tvOS Team Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "tvOS Team Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - Mac Team Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Mac Team Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - Mac Team Ad Hoc Provisioning Profile: *",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Mac Team Ad Hoc Provisioning Profile: *",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - Mac Team Ad Hoc Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Mac Team Ad Hoc Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
		{
			name: "Xcode Managed - Mac Team Provisioning Profile: {bundle id}",
			profile: Profile{
				Attributes: ProfileAttributes{
					Name: "Mac Team Provisioning Profile: {bundle id}",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.profile

			if got := p.isXcodeManaged(); got != tt.want {
				t.Errorf("Profile.isXcodeManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}
