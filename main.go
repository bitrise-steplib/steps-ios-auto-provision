package main

import (
	"os"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-ios-auto-provision/autoprovision"
)

func failf(s string, args ...interface{}) {
	log.Errorf(s, args...)
	os.Exit(1)
}

func main() {
	_, err := autoprovision.ParseConfig()
	if err != nil {
		failf(err.Error())
	}
}
