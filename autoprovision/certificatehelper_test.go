package autoprovision

import (
	"reflect"
	"testing"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-xcode/certificateutil"
	"github.com/bitrise-steplib/steps-ios-auto-provision/appstoreconnect"
)

func mockAPIClient(certs map[CertificateType][]APICertificate) certificateSource {
	return certificateSource{
		queryCertificateBySerialFunc: func(client *appstoreconnect.Client, serial string) ([]APICertificate, error) {
			for _, certList := range certs {
				for _, cert := range certList {
					if cert.Certificate.Serial == serial {
						return []APICertificate{cert}, nil
					}
				}
			}
			return nil, nil
		},
		queryAllCertificatesFunc: func(client *appstoreconnect.Client) (map[CertificateType][]APICertificate, error) {
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
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	devCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	cert, privateKey, err = certificateutil.GenerateTestCertificate(int64(2), teamID, teamName, commonNameIOSDistribution, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	distributionCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", distributionCert)

	type args struct {
		localCertificates        []certificateutil.CertificateInfoModel
		client                   certificateSource
		requiredCertificateTypes map[CertificateType]bool
		typeToName               map[CertificateType]string
		teamID                   string
		logAllCerts              bool
	}
	tests := []struct {
		name    string
		args    args
		want    map[CertificateType][]APICertificate
		wantErr bool
	}{
		{
			name: "dev local; no API; dev required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client:                   mockAPIClient(map[CertificateType][]APICertificate{}),
				requiredCertificateTypes: map[CertificateType]bool{Development: true, Distribution: false},
				typeToName: map[CertificateType]string{
					Development: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "no local; no API; dev+dist requried",
			args: args{
				localCertificates:        []certificateutil.CertificateInfoModel{},
				client:                   mockAPIClient(map[CertificateType][]APICertificate{}),
				requiredCertificateTypes: map[CertificateType]bool{Development: true, Distribution: true},
				typeToName: map[CertificateType]string{
					Development: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "dev local; none API; dev+dist required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client:                   mockAPIClient(map[CertificateType][]APICertificate{}),
				requiredCertificateTypes: map[CertificateType]bool{Development: true, Distribution: true},
				typeToName: map[CertificateType]string{
					Development: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "dev local; dev API; dev required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
				},
				client: mockAPIClient(map[CertificateType][]APICertificate{
					Development: []APICertificate{{
						Certificate: devCert,
						ID:          "apicertid",
					}},
				}),
				requiredCertificateTypes: map[CertificateType]bool{Development: true, Distribution: false},
				typeToName: map[CertificateType]string{
					Development: "iPhone Developer",
				},
				teamID: "",
			},
			want: map[CertificateType][]APICertificate{
				Development: []APICertificate{{
					Certificate: devCert,
					ID:          "apicertid",
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
				client: mockAPIClient(map[CertificateType][]APICertificate{
					Development: []APICertificate{
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
				requiredCertificateTypes: map[CertificateType]bool{Development: true, Distribution: true},
				typeToName: map[CertificateType]string{
					Development: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "dev+dist local; dist API; dev+dist required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					distributionCert,
				},
				client: mockAPIClient(map[CertificateType][]APICertificate{
					Development: []APICertificate{{
						Certificate: devCert,
						ID:          "apicertid_dev",
					}},
				}),
				requiredCertificateTypes: map[CertificateType]bool{
					Development:  true,
					Distribution: true,
				},
				typeToName: map[CertificateType]string{
					Development: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[CertificateType][]APICertificate{},
			wantErr: true,
		},
		{
			name: "dev+dist local; dev+dist API; dev+dist required",
			args: args{
				localCertificates: []certificateutil.CertificateInfoModel{
					devCert,
					distributionCert,
				},
				client: mockAPIClient(map[CertificateType][]APICertificate{
					Development: []APICertificate{
						{
							Certificate: devCert,
							ID:          "dev",
						},
					},
					Distribution: []APICertificate{
						{
							Certificate: distributionCert,
							ID:          "dist",
						},
					},
				}),
				requiredCertificateTypes: map[CertificateType]bool{Development: true, Distribution: true},
				typeToName: map[CertificateType]string{
					Development:  "iPhone Developer",
					Distribution: "iPhone Distribution",
				},
				teamID: "",
			},
			want: map[CertificateType][]APICertificate{
				Development: []APICertificate{{
					Certificate: devCert,
					ID:          "dev",
				}},
				Distribution: []APICertificate{{
					Certificate: distributionCert,
					ID:          "dist",
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValidCertificates(tt.args.localCertificates, tt.args.client, tt.args.requiredCertificateTypes, tt.args.typeToName, tt.args.teamID, tt.args.logAllCerts)
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

func Test_logUpdatedAPICertificates(t *testing.T) {
	log.SetEnableDebugLog(true)

	const teamID = "MYTEAMID"
	const commonName = "iPhone Developer: test"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	serial := int64(1234)

	certs := []certificateutil.CertificateInfoModel{}
	for i := 1; i <= 4; i++ {
		cert, privateKey, err := certificateutil.GenerateTestCertificate(serial, teamID, teamName, commonName, time.Now().AddDate(0, 0, i))
		if err != nil {
			t.Errorf("init: failed to generate certificate, error: %s", err)
		}
		certInfo := certificateutil.NewCertificateInfo(*cert, privateKey)
		t.Logf("Test certificate generated. %s", certInfo)

		certs = append(certs, certInfo)
	}

	mapConnect := func(certs []certificateutil.CertificateInfoModel) []APICertificate {
		var connectCerts []APICertificate
		for i, c := range certs {
			connectCerts = append(connectCerts, APICertificate{
				Certificate: c,
				ID:          string(i),
			})
		}
		return connectCerts
	}

	tests := []struct {
		name              string
		localCertificates []certificateutil.CertificateInfoModel
		APICertificates   []APICertificate
		want              bool
	}{
		{
			name:              "no newer",
			localCertificates: certs[:1],
			APICertificates:   mapConnect(certs[:1]),
			want:              false,
		},
		{
			name:              "one newer",
			localCertificates: certs[:1],
			APICertificates:   mapConnect(certs[:2]),
			want:              true,
		},
		{
			name:              "two newer",
			localCertificates: certs[:1],
			APICertificates:   mapConnect(certs[:3]),
			want:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := logUpdatedAPICertificates(tt.localCertificates, tt.APICertificates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchLocalCertificatesToConnectCertificates() = %v, want %v", got, tt.want)
			}
		})
	}
}
