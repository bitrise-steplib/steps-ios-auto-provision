package autoprovision

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

func mockAPIClient(certs map[appstoreconnect.CertificateType][]APICertificate) CertificateSource {
	return CertificateSource{
		queryCertificateBySerialFunc: func(client *appstoreconnect.Client, serial *big.Int) (APICertificate, error) {
			for _, certList := range certs {
				for _, cert := range certList {
					if cert.Certificate.Certificate.SerialNumber == serial {
						return cert, nil
					}
				}
			}
			return APICertificate{}, fmt.Errorf("certificate with serial %s not found", serial.String())
		},
		queryAllCertificatesFunc: func(client *appstoreconnect.Client) (map[appstoreconnect.CertificateType][]APICertificate, error) {
			return certs, nil
		},
	}
}

func TestGetValidCertificates(t *testing.T) {
	log.SetEnableDebugLog(true)

	const teamID = "MYTEAMID"
	// Could be "Apple Development: test"
	const commonNameIOSDevelopment = "iPhone Developer: test"
	// Could be "Apple Distribution: test"
	const commonNameIOSDistribution = "iPhone Distribution: test"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	expiry := time.Now().AddDate(1, 0, 0)

	cert, privateKey, err := certificateutil.GenerateTestCertificate(int64(1), teamID, teamName, commonNameIOSDevelopment, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate: %s", err)
	}
	devCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	cert, privateKey, err = certificateutil.GenerateTestCertificate(int64(2), teamID, teamName, "iPhone Developer: test2", expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate: %s", err)
	}
	devCert2 := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	distCert, privateKey, err := certificateutil.GenerateTestCertificate(int64(10), teamID, teamName, commonNameIOSDistribution, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate: %s", err)
	}
	distributionCert := certificateutil.NewCertificateInfo(*distCert, privateKey)
	t.Logf("Test certificate generated. %s", distributionCert)

	type args struct {
		localCertificates        []certificateutil.CertificateInfoModel
		client                   CertificateSource
		requiredCertificateTypes map[appstoreconnect.CertificateType]bool
		teamID                   string
	}
	tests := []struct {
		name    string
		args    args
		want    map[appstoreconnect.CertificateType][]APICertificate
		wantErr bool
	}{
		{
			name: "dev local; no API; dev required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client:                   mockAPIClient(map[appstoreconnect.CertificateType][]APICertificate{}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
				teamID:                   "",
			},
			want:    map[appstoreconnect.CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "2 dev local with same name; 1 dev API; dev required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					devCert,
					devCert2,
				},
				client: mockAPIClient(map[appstoreconnect.CertificateType][]APICertificate{
					appstoreconnect.IOSDevelopment: []APICertificate{{
						Certificate: devCert,
						ID:          "devcert",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
				teamID:                   "",
			},
			want: map[appstoreconnect.CertificateType][]APICertificate{
				appstoreconnect.IOSDevelopment: []APICertificate{{
					Certificate: devCert,
					ID:          "devcert",
				}},
			},
			wantErr: false,
		},
		{
			name: "no local; no API; dev+dist requried",
			args: args{
				localCertificates:        []certificateutil.CertificateInfoModel{},
				client:                   mockAPIClient(map[appstoreconnect.CertificateType][]APICertificate{}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
				teamID:                   "",
			},
			want:    map[appstoreconnect.CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "dev local; none API; dev+dist required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client:                   mockAPIClient(map[appstoreconnect.CertificateType][]APICertificate{}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
				teamID:                   "",
			},
			want:    map[appstoreconnect.CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "dev local; dev API; dev required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client: mockAPIClient(map[appstoreconnect.CertificateType][]APICertificate{
					appstoreconnect.IOSDevelopment: []APICertificate{{
						Certificate: devCert,
						ID:          "apicertid",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
				teamID:                   "",
			},
			want: map[appstoreconnect.CertificateType][]APICertificate{
				appstoreconnect.IOSDevelopment: []APICertificate{{
					Certificate: devCert,
					ID:          "apicertid",
				}},
			},
			wantErr: false,
		},
		{
			name: "2 dev local; 1 dev API; dev required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					devCert2,
				},
				client: mockAPIClient(map[appstoreconnect.CertificateType][]APICertificate{
					appstoreconnect.IOSDevelopment: []APICertificate{{
						Certificate: devCert,
						ID:          "dev1",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: false},
				teamID:                   "",
			},
			want: map[appstoreconnect.CertificateType][]APICertificate{
				appstoreconnect.IOSDevelopment: []APICertificate{{
					Certificate: devCert,
					ID:          "dev1",
				}},
			},
			wantErr: false,
		},
		{
			name: "dev local; dev+dist API; both required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client: mockAPIClient(map[appstoreconnect.CertificateType][]APICertificate{
					appstoreconnect.IOSDevelopment: []APICertificate{
						{
							Certificate: devCert,
							ID:          "apicertid_dev",
						},
						{
							Certificate: distributionCert,
							ID:          "apicertid_dist",
						},
					},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
				teamID:                   "",
			},
			want:    map[appstoreconnect.CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "dev+dist local; dist API; dev+dist required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					distributionCert,
				},
				client: mockAPIClient(map[appstoreconnect.CertificateType][]APICertificate{
					appstoreconnect.IOSDevelopment: []APICertificate{{
						Certificate: devCert,
						ID:          "dev",
					}},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{
					appstoreconnect.IOSDevelopment:  true,
					appstoreconnect.IOSDistribution: true,
				},
				teamID: "",
			},
			want:    map[appstoreconnect.CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "dev+dist local; dev+dist API; dev+dist required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					distributionCert,
				},
				client: mockAPIClient(map[appstoreconnect.CertificateType][]APICertificate{
					appstoreconnect.IOSDevelopment: []APICertificate{
						{
							Certificate: devCert,
							ID:          "dev",
						},
					},
					appstoreconnect.IOSDistribution: []APICertificate{
						{
							Certificate: distributionCert,
							ID:          "dist",
						},
					},
				}),
				requiredCertificateTypes: map[appstoreconnect.CertificateType]bool{appstoreconnect.IOSDevelopment: true, appstoreconnect.IOSDistribution: true},
				teamID:                   "",
			},
			want: map[appstoreconnect.CertificateType][]APICertificate{
				appstoreconnect.IOSDevelopment: []APICertificate{{
					Certificate: devCert,
					ID:          "dev",
				}},
				appstoreconnect.IOSDistribution: []APICertificate{{
					Certificate: distributionCert,
					ID:          "dist",
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValidCertificates(tt.args.localCertificates, tt.args.client, tt.args.requiredCertificateTypes, tt.args.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValidCertificates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for certType, wantCerts := range tt.want {
				if !reflect.DeepEqual(wantCerts, got[certType]) {
					t.Errorf("GetValidCertificates()[%s] = %v, want %v", certType, got, tt.want)
				}
			}
		})
	}
}
