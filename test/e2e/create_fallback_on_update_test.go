package e2e

import (
	"strings"
	"testing"

	uitest "github.com/cppforlife/go-cli-ui/ui/test"
)

func TestCreateFallbackOnUpdate(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	kapp := Kapp{t, env.Namespace, env.KappBinaryPath, logger}

	objNs := env.Namespace + "-create-fallback-on-update"
	yaml1 := strings.Replace(`
---
apiVersion: v1
kind: Namespace
metadata:
  name: __ns__
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
  namespace: __ns__
  annotations:
    kapp.k14s.io/create-strategy: fallback-on-update
imagePullSecrets:
- name: pull-secret
`, "__ns__", objNs, -1)

	name := "test-create-fallback-on-update"
	cleanUp := func() {
		kapp.RunWithOpts([]string{"delete", "-a", name}, RunOpts{AllowError: true})
	}

	cleanUp()
	defer cleanUp()

	logger.Section("deploy expecting service account creation to fail", func() {
		yamlNoCreateStrategy := strings.Replace(yaml1, "create-strategy", "create-strategy.xxx", -1)

		_, err := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{AllowError: true, StdinReader: strings.NewReader(yamlNoCreateStrategy)})

		if !strings.Contains(err.Error(), `serviceaccounts "default" already exists`) {
			t.Fatalf("Expected serviceaccount to be already created, but error was: %s", err)
		}

		cleanUp()
	})

	logger.Section("deploy with create-strategy annotation", func() {
		kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name},
			RunOpts{StdinReader: strings.NewReader(yaml1)})
	})

	logger.Section("deploy second time with expected no changes", func() {
		out, _ := kapp.RunWithOpts([]string{"deploy", "-f", "-", "-a", name, "--json"},
			RunOpts{StdinReader: strings.NewReader(yaml1)})

		resp := uitest.JSONUIFromBytes(t, []byte(out))

		if len(resp.Tables[0].Rows) != 0 {
			t.Fatalf("Expected to see no changes, but did not: '%s'", out)
		}
	})
}