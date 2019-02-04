package main

import (
	"os"

	"github.com/bitrise-io/go-utils/log"
)

func failf(s string, args ...interface{}) {
	log.Errorf(s, args...)
	os.Exit(1)
}

func main() {
	config, err := ParseConfig()
	if err != nil {
		failf(err.Error())
	}

}
