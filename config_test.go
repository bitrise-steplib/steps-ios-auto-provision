package main

import (
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func Test_splitByPipe(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want []string
	}{
		{
			name: "empty",
			arg:  "",
			want: []string(nil),
		},
		{
			name: "pipe char",
			arg:  "|",
			want: []string(nil),
		},
		{
			name: "space + pipe char",
			arg:  " |",
			want: []string(nil),
		},
		{
			name: "pipe char + space",
			arg:  "| ",
			want: []string(nil),
		},
		{
			name: "space + pipe char + spaces",
			arg:  " |  ",
			want: []string(nil),
		},
		{
			name: "newlines + pipe char + newline",
			arg:  "\n\n|\n",
			want: []string(nil),
		},
		{
			name: "newline + pipe char",
			arg:  `|`,
			want: []string(nil),
		},
		{
			name: "single element",
			arg:  `url`,
			want: []string{"url"},
		},
		{
			name: "multiple elements",
			arg:  `url1|url2|url3`,
			want: []string{"url1", "url2", "url3"},
		},
		{
			name: "multiple elements with spaces and newlines",
			arg: `url1
|url2   |

url3`,
			want: []string{"url1", "url2", "url3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotItems := splitByPipe(tt.arg, true); !reflect.DeepEqual(gotItems, tt.want) {
				t.Errorf("splitByPipe() = %v, want %v", gotItems, tt.want)
			}
		})
	}
}

// ParseConfig test is implemented by mocking a subprocess,
// this idea is seen in the standard library tests: https://golang.org/src/os/exec/exec_test.go
// Other way would be to pass the environment to stepconf.Parse(),
// but implementing os.Getenv() is tricky: https://golang.org/src/os/env.go?s=2860:2890#L91
func TestParseConfig(t *testing.T) {
	if os.Getenv("TEST_PARSE_CONFIG") == "1" {
		configErr := `failed to parse config:
- ProjectPath: file does not exist
- Scheme: required variable is not present
- DistributionType: required variable is not present
- GenerateProfiles: value is not in value options (opt[no,yes])
- VerboseLog: value is not in value options (opt[no,yes])
- certificateURLList: required variable is not present
- certificatePassphraseList: required variable is not present
- KeychainPath: required variable is not present
- KeychainPassword: required variable is not present
- BuildURL: required variable is not present
- BuildAPIToken: required variable is not present`
		_, err := ParseConfig()
		if !reflect.DeepEqual(err.Error(), configErr) {
			t.Errorf("ParseConfig() = %v, want %v", err.Error(), configErr)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestParseConfig")
	cmd.Env = []string{"TEST_PARSE_CONFIG=1"}
	out, err := cmd.Output()
	if err != nil {
		t.Errorf(string(out))
	}
}

func TestConfig_ValidateCertificateAndPassphraseCount(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name:    "",
			config:  Config{certificateURLList: "url", certificatePassphraseList: "pass"},
			wantErr: "",
		},
		{
			name:    "",
			config:  Config{certificateURLList: "url1|url2", certificatePassphraseList: "pass1|"},
			wantErr: "",
		},
		{
			name:    "",
			config:  Config{certificateURLList: "url1|url2", certificatePassphraseList: "pass1"},
			wantErr: "certificates count (2) and passphrases count (1) should match",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.ValidateCertificateAndPassphraseCount(); (len(tt.wantErr) > 0 && err == nil) || (len(tt.wantErr) > 0 && err.Error() != tt.wantErr) || (len(tt.wantErr) == 0 && err != nil) {
				t.Errorf("Config.ValidateCertificateAndPassphraseCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
