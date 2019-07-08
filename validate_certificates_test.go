package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-xcode/certificateutil"
)

func TestGetMatchingCertificates(t *testing.T) {
	log.SetEnableDebugLog(true)

	const teamID = "MYTEAMID"
	const commonNameIOSDevelopment = "iPhone Developer: test"
	const commonNameAppleDevelopment = "Apple Development: test"
	const commonNameIOSDistribution = "iPhone Distribution: test"
	const commonNameAppleDistribution = "Apple Distribution: test"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	expiry := time.Now().AddDate(1, 0, 0)
	serial := int64(1234)

	cert, privateKey, err := generateTestCertificate(serial, teamID, teamName, commonNameIOSDevelopment, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	devCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	cert, privateKey, err = generateTestCertificate(serial, teamID, teamName, commonNameIOSDistribution, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	distributionCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", distributionCert)

	type args struct {
		certificates                []certificateutil.CertificateInfoModel
		AppStoreConnectCertificates map[CertificateType][]AppStoreConnectCertificate
		distribution                ProfileType
		typeToName                  map[CertificateType]string
		teamID                      string
	}
	tests := []struct {
		name    string
		args    args
		want    map[CertificateType][]AppStoreConnectCertificate
		wantErr bool
	}{
		{
			name: "one local cert, not found on App Store Connect",
			args: args{
				certificates:                []certificateutil.CertificateInfoModel{devCert},
				AppStoreConnectCertificates: map[CertificateType][]AppStoreConnectCertificate{},
				distribution:                Development,
				typeToName: map[CertificateType]string{
					DevelopmentCertificate: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "no local certificates",
			args: args{
				certificates:                []certificateutil.CertificateInfoModel{},
				AppStoreConnectCertificates: map[CertificateType][]AppStoreConnectCertificate{},
				distribution:                Development,
				typeToName: map[CertificateType]string{
					DevelopmentCertificate: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "App store distribution but only development local certificate present",
			args: args{
				certificates:                []certificateutil.CertificateInfoModel{devCert},
				AppStoreConnectCertificates: map[CertificateType][]AppStoreConnectCertificate{},
				distribution:                AppStore,
				typeToName: map[CertificateType]string{
					DevelopmentCertificate: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "Development distribution local and app store cert present.",
			args: args{
				certificates: []certificateutil.CertificateInfoModel{devCert},
				AppStoreConnectCertificates: map[CertificateType][]AppStoreConnectCertificate{
					DevelopmentCertificate: []AppStoreConnectCertificate{{
						certificate:       devCert,
						appStoreConnectID: "apicertid",
					}},
				},
				distribution: Development,
				typeToName: map[CertificateType]string{
					DevelopmentCertificate: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]AppStoreConnectCertificate{},
			wantErr: false,
		},
		{
			name: "App Store distribution local and app store dev cert present, distribution only on App Store Connect.",
			args: args{
				certificates: []certificateutil.CertificateInfoModel{devCert},
				AppStoreConnectCertificates: map[CertificateType][]AppStoreConnectCertificate{
					DevelopmentCertificate: []AppStoreConnectCertificate{
						{
							certificate:       devCert,
							appStoreConnectID: "apicertid_dev",
						},
						{
							certificate:       distributionCert,
							appStoreConnectID: "apicertid_dist",
						},
					},
				},
				distribution: AppStore,
				typeToName: map[CertificateType]string{
					DevelopmentCertificate: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "App Store distribution local and app store dev cert present, distribution only local.",
			args: args{
				certificates: []certificateutil.CertificateInfoModel{devCert, distributionCert},
				AppStoreConnectCertificates: map[CertificateType][]AppStoreConnectCertificate{
					DevelopmentCertificate: []AppStoreConnectCertificate{
						{
							certificate:       devCert,
							appStoreConnectID: "apicertid_dev",
						},
					},
				},
				distribution: AppStore,
				typeToName: map[CertificateType]string{
					DevelopmentCertificate: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "App Store distribution local and app store dev and distribution cert present.",
			args: args{
				certificates: []certificateutil.CertificateInfoModel{devCert, distributionCert},
				AppStoreConnectCertificates: map[CertificateType][]AppStoreConnectCertificate{
					DevelopmentCertificate: []AppStoreConnectCertificate{
						{
							certificate:       devCert,
							appStoreConnectID: "apicertid_dev",
						},
					},
					DistributionCertificate: []AppStoreConnectCertificate{
						{
							certificate:       distributionCert,
							appStoreConnectID: "apicertid_dist",
						},
					},
				},
				distribution: AppStore,
				typeToName: map[CertificateType]string{
					DevelopmentCertificate:  "iPhone Developer",
					DistributionCertificate: "iPhone Distribution",
				},
				teamID: "",
			},
			want:    map[CertificateType][]AppStoreConnectCertificate{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMatchingCertificates(tt.args.certificates, tt.args.AppStoreConnectCertificates, tt.args.distribution, tt.args.typeToName, tt.args.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMatchingCertificates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for certType, wantCerts := range tt.want {
				if !reflect.DeepEqual(wantCerts, got[certType]) {
					t.Errorf("GetMatchingCertificates()[%s] = %v, want %v", certType, got, tt.want)
				}
			}
		})
	}
}
