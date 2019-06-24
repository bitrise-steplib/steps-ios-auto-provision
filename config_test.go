package main

import (
	"reflect"
	"testing"
)

func Test_split(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		sep  string
		want []string
	}{
		{
			name: "empty",
			arg:  "",
			sep:  "|",
			want: []string(nil),
		},
		{
			name: "pipe char",
			arg:  "|",
			sep:  "|",
			want: []string(nil),
		},
		{
			name: "space + pipe char",
			arg:  " |",
			sep:  "|",
			want: []string(nil),
		},
		{
			name: "pipe char + space",
			arg:  "| ",
			sep:  "|",
			want: []string(nil),
		},
		{
			name: "space + pipe char + spaces",
			arg:  " |  ",
			sep:  "|",
			want: []string(nil),
		},
		{
			name: "newlines + pipe char + newline",
			arg:  "\n\n|\n",
			sep:  "|",
			want: []string(nil),
		},
		{
			name: "newline + pipe char",
			arg:  `|`,
			sep:  "|",
			want: []string(nil),
		},
		{
			name: "single element",
			arg:  `url`,
			sep:  "|",
			want: []string{"url"},
		},
		{
			name: "multiple elements",
			arg:  `url1|url2|url3`,
			sep:  "|",
			want: []string{"url1", "url2", "url3"},
		},
		{
			name: "multiple elements with spaces and newlines",
			arg: `url1
|url2   |

url3`,
			sep:  "|",
			want: []string{"url1", "url2", "url3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotItems := split(tt.arg, tt.sep, true); !reflect.DeepEqual(gotItems, tt.want) {
				t.Errorf("splitByPipe() = %v, want %v", gotItems, tt.want)
			}
		})
	}
}

func TestConfig_ValidateCertificates(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		want, want1 []string
		wantErr     string
	}{
		{
			name:    "",
			config:  Config{CertificateURLList: "url", CertificatePassphraseList: "pass"},
			want:    []string{"url"},
			want1:   []string{"pass"},
			wantErr: "",
		},
		{
			name:    "",
			config:  Config{CertificateURLList: "url1|url2", CertificatePassphraseList: "pass1|"},
			want:    []string{"url1", "url2"},
			want1:   []string{"pass1", ""},
			wantErr: "",
		},
		{
			name:    "",
			config:  Config{CertificateURLList: "url1|url2", CertificatePassphraseList: "pass1"},
			want:    nil,
			want1:   nil,
			wantErr: "certificates count (2) and passphrases count (1) should match",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.config.ValidateCertificates()
			if (len(tt.wantErr) > 0 && err == nil) || (len(tt.wantErr) > 0 && err.Error() != tt.wantErr) || (len(tt.wantErr) == 0 && err != nil) {
				t.Errorf("Config.ValidateCertificateAndPassphraseCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.ValidateCertificates() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Config.ValidateCertificates() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
