package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

//go:embed templates/istioctl.yaml
var installable string

func extractIstioctlInfo(version, os, arch string) (string, error) {
	out := fmt.Sprintf("/tmp/istioctl-%s-%s-%s.tar.gz", version, os, arch)
	if os == "osx" && arch == "amd64" {
		arch = ""
	} else {
		arch = "-" + arch
	}
	src := fmt.Sprintf("https://github.com/istio/istio/releases/download/%s/istioctl-%s-%s%s.tar.gz", version, version, os, arch)
	fmt.Println("fetching", src, "...")
	if err := sh.RunV("curl", "-sSLf", "-o", out, src); err != nil {
		return "", nil
	}
	sha, err := sh.Output("sha256sum", out)
	if err != nil {
		return "", err
	}
	return strings.Fields(sha)[0], nil
}

func generateInstallable(istioVersion, dist string) (string, error) {
	m := map[string]string{}
	m["istioVersion"] = istioVersion

	for _, os := range []string{"osx", "linux"} {
		for _, arch := range []string{"arm64", "amd64"} {
			sha, err := extractIstioctlInfo(istioVersion, os, arch)
			if err != nil {
				return "", err
			}
			m[os+arch] = sha
		}
	}

	dir := filepath.Join(dist, "installables", "istioctl")
	_ = os.MkdirAll(dir, os.ModePerm)
	f, _ := os.Create(filepath.Join(dir, istioVersion+".yaml"))
	t := template.Must(template.New("").Parse(installable))
	t.Execute(f, m)
	return "", nil
}
