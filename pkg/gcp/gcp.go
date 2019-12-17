package gcp

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

var gcloudPath = "/usr/lib/google-cloud-sdk/bin/gcloud"

func AuthServiceAccount() error {
	var stderr bytes.Buffer
	path := os.Getenv("GOOGLE_KEY_PATH")
	if len(path) == 0 {
		return nil
	}

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Errorf("Failed to access %v err=%v", path, err)
	}

	args := []string{"auth", "activate-service-account", "--key-file", path}
	cmd := exec.Command(gcloudPath, args...)
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		return errors.Errorf("error activating gcp service account cmd=%v err=%v output=%s stderr=%s",
			strings.Join(append([]string{gcloudPath}, args...), " "),
			err,
			output,
			string(stderr.Bytes()))
	}
	return nil
}
