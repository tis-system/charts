package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/magefile/mage/sh"
	"k8s.io/utils/strings/slices"
)

var repos = []string{
	"install-cni",
	"istioctl",
	"operator",
	"pilot",
	"proxyv2",
	"ztunnel",
}

type catalog struct {
	Repositories []string `json:"repositories"`
}

type tagged struct {
	Tags []string `json:"tags"`
}

func getFIPSRepos() ([]string, error) {
	token := os.Getenv("TOKEN")
	result, err := sh.Output("curl", "-sSLf", "-u", ":"+token, "https://docker.cloudsmith.io/v2/tetrate/tid-fips-containers/_catalog")
	if err != nil {
		return nil, err
	}
	var c catalog
	err = json.Unmarshal([]byte(result), &c)
	if err != nil {
		return nil, err
	}

	return c.Repositories, nil
}

func getTaggedFIPSImage(repo, version, token string) error {
	result, err := sh.Output("curl", "-sSLf", "-u", ":"+token, "https://docker.cloudsmith.io/v2/tetrate/tid-fips-containers/"+repo+"/tags/list")
	if err != nil {
		return err
	}

	var t tagged
	err = json.Unmarshal([]byte(result), &t)
	if err != nil {
		return err
	}
	if slices.Contains(t.Tags, version+"-tetratefips-v0") {
		return nil
	}
	return errors.New("not found")
}

func getAvailableFIPSImagesForVersion(repos []string, version string) ([]string, error) {
	token := os.Getenv("TOKEN")
	var included []string
	for _, repo := range repos {
		err := getTaggedFIPSImage(repo, version, token)
		if err == nil {
			included = append(included, repo)
		}
	}
	return included, nil
}
