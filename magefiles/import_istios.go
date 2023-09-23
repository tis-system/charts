package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/dio/magex/tool"
	"github.com/magefile/mage/sh"
	"golang.org/x/sync/errgroup"
)

func importIstioCharts(ctx context.Context, istios []istio) error {
	g, _ := errgroup.WithContext(ctx)
	for _, entry := range istios {
		entry := entry
		fmt.Println("importing", entry.Version, "...")
		g.TryGo(func() error {
			return importIstioChartsPerVersion(ctx, entry.Version, entry.Tetrate)
		})
	}
	return g.Wait()
}

func importIstioChartsPerVersion(ctx context.Context, version, tetrateV string) error {
	if len(tetrateV) == 0 {
		tetrateV = "0"
	}
	fmt.Println("generate installable for", version)
	generateInstallable(version, dist())

	name := "istio-" + version + "-tetrate-v" + tetrateV
	targz := name + "-linux-amd64.tar.gz"
	downloads := "downloads"
	downloaded := path.Join(downloads, name+".tar.gz")

	_ = os.MkdirAll(downloads, os.ModePerm)
	err := sh.Run("curl", "-sSLo", downloaded, "https://dl.getistio.io/public/raw/files/"+targz)
	if err != nil {
		return err
	}

	err = sh.Run("tar", "xzf", downloaded, "-C", downloads)
	if err != nil {
		return err
	}

	for _, versioned := range []struct {
		name          string
		versionSuffix string
		patch         func(string, string) error
	}{
		{"istio", "", func(name, version string) error {
			// This is from: https://github.com/tetratelabs/helm-charts/blob/77b0b41dd6b9b2765fb7e5279e1dd5d2dd4598f0/.github/workflows/charts.yml#L40-L44.
			return sh.Run("modifier/fluxcd-v1/modify.sh", name, version)
		}},
	} {
		build := path.Join("charts", versioned.name, version)
		dist := path.Join(dist(), versioned.name, version)

		_ = os.RemoveAll(build)
		_ = os.MkdirAll(build, os.ModePerm)
		err = sh.Run("cp", "-fa", path.Join(downloads, name, "manifests", "charts")+"/.", build)
		if err != nil {
			return err
		}
		entries, err := os.ReadDir(path.Join(build))
		if err != nil {
			return err
		}
		charts := []string{}
		for _, entry := range entries {
			entry := entry
			if entry.IsDir() {
				chart := path.Join(build, entry.Name(), "Chart.yaml")
				_, err := os.Stat(chart)
				if err != nil {
					// If missing we need to check if this directory has subdirectories.
					subentries, err := os.ReadDir(filepath.Dir(chart))
					if err != nil {
						return err
					}
					for _, subentry := range subentries {
						subchart := path.Join(filepath.Dir(chart), subentry.Name(), "Chart.yaml")
						_, err := os.Stat(subchart)
						if err != nil {
							return err
						}
						charts = append(charts, filepath.Dir(subchart))
					}
					continue
				}
				charts = append(charts, filepath.Dir(chart))
			}
		}

		if versioned.patch != nil {
			err = versioned.patch(versioned.name, version)
			if err != nil {
				return err
			}
		}

		_ = os.RemoveAll(dist)
		helmArgs := []string{
			"package",
			"--destination", dist,
			"--version", version + versioned.versionSuffix,
		}
		err = toolbox().RunWith(ctx, tool.RunWithOption{Env: map[string]string{
			"TZ": "UTC",
		}}, "helm", append(helmArgs, charts...)...)
		if err != nil {
			return err
		}

		// Note: This requires gnu-tar (tar (GNU tar) 1.35).
		// On macOS, brew install gnu-tar.
		// export PATH="/opt/homebrew/opt/gnu-tar/libexec/gnubin:$PATH"
		infoTime, err := getCreationTime(downloaded)
		if err != nil {
			return err
		}
		err = modifyTimestamp(dist, infoTime)
		if err != nil {
			return err
		}
	}
	return nil
}
