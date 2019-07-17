package autoprovision

/*
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

	cert, privateKey, err := certificateutil.GenerateTestCertificate(serial, teamID, teamName, commonNameIOSDevelopment, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	devCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", devCert)

	cert, privateKey, err = certificateutil.GenerateTestCertificate(serial, teamID, teamName, commonNameIOSDistribution, expiry)
	if err != nil {
		t.Errorf("init: failed to generate certificate, error: %s", err)
	}
	distributionCert := certificateutil.NewCertificateInfo(*cert, privateKey)
	t.Logf("Test certificate generated. %s", distributionCert)

	type args struct {
		certificates                []certificateutil.CertificateInfoModel
		AppStoreConnectCertificates map[appstoreconnect.CertificateType][]AppStoreConnectCertificate
		requiredCertificatetypes    []appstoreconnect.CertificateType
		typeToName                  map[appstoreconnect.CertificateType]string
		teamID                      string
	}
	tests := []struct {
		name    string
		args    args
		want    map[appstoreconnect.CertificateType][]AppStoreConnectCertificate
		wantErr bool
	}{
		{
			name: "one local cert, not found on App Store Connect",
			args: args{
				certificates:                []certificateutil.CertificateInfoModel{devCert},
				AppStoreConnectCertificates: map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
				requiredCertificatetypes:    []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment},
				typeToName: map[appstoreconnect.CertificateType]string{
					appstoreconnect.IOSDevelopment: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "no local certificates",
			args: args{
				certificates:                []certificateutil.CertificateInfoModel{},
				AppStoreConnectCertificates: map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
				requiredCertificatetypes:    []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution},
				typeToName: map[appstoreconnect.CertificateType]string{
					appstoreconnect.IOSDevelopment: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "App store distribution but only development local certificate present",
			args: args{
				certificates:                []certificateutil.CertificateInfoModel{devCert},
				AppStoreConnectCertificates: map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
				requiredCertificatetypes:    []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution},
				typeToName: map[appstoreconnect.CertificateType]string{
					appstoreconnect.IOSDevelopment: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "Development distribution local and app store cert present.",
			args: args{
				certificates: []certificateutil.CertificateInfoModel{devCert},
				AppStoreConnectCertificates: map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{
					appstoreconnect.IOSDevelopment: []AppStoreConnectCertificate{{
						certificate:       devCert,
						appStoreConnectID: "apicertid",
					}},
				},
				requiredCertificatetypes: []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment},
				typeToName: map[appstoreconnect.CertificateType]string{
					appstoreconnect.IOSDevelopment: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
			wantErr: false,
		},
		{
			name: "App Store distribution local and app store dev cert present, distribution only on App Store Connect.",
			args: args{
				certificates: []certificateutil.CertificateInfoModel{devCert},
				AppStoreConnectCertificates: map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{
					appstoreconnect.IOSDevelopment: []AppStoreConnectCertificate{
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
				requiredCertificatetypes: []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution},
				typeToName: map[appstoreconnect.CertificateType]string{
					appstoreconnect.IOSDevelopment: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "App Store distribution local and app store dev cert present, distribution only local.",
			args: args{
				certificates: []certificateutil.CertificateInfoModel{devCert, distributionCert},
				AppStoreConnectCertificates: map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{
					appstoreconnect.IOSDevelopment: []AppStoreConnectCertificate{
						{
							certificate:       devCert,
							appStoreConnectID: "apicertid_dev",
						},
					},
				},
				requiredCertificatetypes: []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution},
				typeToName: map[appstoreconnect.CertificateType]string{
					appstoreconnect.IOSDevelopment: "iPhone Developer",
				},
				teamID: "",
			},
			want:    map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
			wantErr: true,
		},
		{
			name: "App Store distribution local and app store dev and distribution cert present.",
			args: args{
				certificates: []certificateutil.CertificateInfoModel{devCert, distributionCert},
				AppStoreConnectCertificates: map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{
					appstoreconnect.IOSDevelopment: []AppStoreConnectCertificate{
						{
							certificate:       devCert,
							appStoreConnectID: "apicertid_dev",
						},
					},
					appstoreconnect.IOSDistribution: []AppStoreConnectCertificate{
						{
							certificate:       distributionCert,
							appStoreConnectID: "apicertid_dist",
						},
					},
				},
				requiredCertificatetypes: []appstoreconnect.CertificateType{appstoreconnect.IOSDevelopment, appstoreconnect.IOSDistribution},
				typeToName: map[appstoreconnect.CertificateType]string{
					appstoreconnect.IOSDevelopment:  "iPhone Developer",
					appstoreconnect.IOSDistribution: "iPhone Distribution",
				},
				teamID: "",
			},
			want:    map[appstoreconnect.CertificateType][]AppStoreConnectCertificate{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMatchingCertificates(tt.args.certificates, tt.args.AppStoreConnectCertificates, tt.args.requiredCertificatetypes, tt.args.typeToName, tt.args.teamID)
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

func Test_matchLocalCertificatesToConnectCertificates(t *testing.T) {
	log.SetEnableDebugLog(true)

	const teamID = "MYTEAMID"
	const commonName = "iPhone Developer: test"
	const teamName = "BITFALL FEJLESZTO KORLATOLT FELELOSSEGU TARSASAG"
	serial := int64(1234)

	certs := []certificateutil.CertificateInfoModel{}
	for i := 1; i < 4; i++ {
		cert, privateKey, err := certificateutil.GenerateTestCertificate(serial, teamID, teamName, commonName, time.Now().AddDate(0, 0, i))
		if err != nil {
			t.Errorf("init: failed to generate certificate, error: %s", err)
		}
		certInfo := certificateutil.NewCertificateInfo(*cert, privateKey)
		t.Logf("Test certificate generated. %s", certInfo)

		certs = append(certs, certInfo)
	}

	mapConnect := func(certs []certificateutil.CertificateInfoModel) []AppStoreConnectCertificate {
		var connectCerts []AppStoreConnectCertificate
		for i, c := range certs {
			connectCerts = append(connectCerts, AppStoreConnectCertificate{
				certificate:       c,
				appStoreConnectID: string(i),
			})
		}
		return connectCerts
	}

	tests := []struct {
		name                string
		localCertificates   []certificateutil.CertificateInfoModel
		connectCertificates []AppStoreConnectCertificate
		want                []AppStoreConnectCertificate
	}{
		{
			name:                "no newer",
			localCertificates:   certs[:0],
			connectCertificates: mapConnect(certs[:0]),
			want:                mapConnect(certs[:0]),
		},
		{
			name:                "one newer",
			localCertificates:   certs[:0],
			connectCertificates: mapConnect(certs[:1]),
			want:                nil,
		},
		{
			name:                "two newer",
			localCertificates:   certs[:0],
			connectCertificates: mapConnect(certs[:2]),
			want:                nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchLocalCertificatesToConnectCertificates(tt.localCertificates, tt.connectCertificates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchLocalCertificatesToConnectCertificates() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
