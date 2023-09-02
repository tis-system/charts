package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dio/magex/tool"
	"github.com/magefile/mage/sh"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/tis-system/charts/pkg/helm/repo"
)

var box *tool.Box

func toolbox() *tool.Box {
	if box == nil {
		box = tool.MustLoadDefault()
	}
	return box
}

func Import(ctx context.Context) error {
	err := toolbox().InstallAll(ctx)
	if err != nil {
		return err
	}

	b, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}
	var loaded config

	err = yaml.Unmarshal(b, &loaded)
	if err != nil {
		return err
	}

	err = importIstioCharts(ctx, loaded.Istios)
	if err != nil {
		return err
	}

	err = importAddonsCharts(ctx, loaded.Addons, loaded.Dependencies)
	if err != nil {
		return err
	}

	indexFile, err := repo.IndexDirectory(path.Join(dist(), "charts"), "")
	if err != nil {
		return err
	}

	return indexFile.WriteFile(path.Join(dist(), "charts", "index.yaml"), os.ModePerm)
}

func importAddonsCharts(ctx context.Context, names []string, deps []dependency) error {
	original := path.Join("downloads", "tetratelabs", "helm-charts")
	_ = os.RemoveAll(original)
	_ = os.MkdirAll(original, os.ModePerm)
	err := sh.Run("git", "clone", "--depth", "1", "--single-branch", "-b", "main", "https://github.com/tetratelabs/helm-charts.git", original)
	if err != nil {
		return err
	}
	charts := []string{}
	for _, name := range names {
		chart := path.Join("charts", "addons", name)
		_ = os.MkdirAll(filepath.Dir(chart), os.ModePerm)

		// Inspect the chart, and create a directory under addons/name.
		c := path.Join(original, "charts", name, "Chart.yaml")
		v, err := getChartVersion(c)
		if err != nil {
			return err
		}
		dst := path.Join(filepath.Dir(chart), name, v)
		_ = os.MkdirAll(dst, os.ModePerm)
		err = sh.Run("cp", "-af", path.Join(original, "charts", name)+"/.", dst)
		if err != nil {
			return err
		}
		charts = append(charts, dst)
	}

	for _, dep := range deps {
		err = toolbox().Run(ctx, "helm", "repo", "add", dep.Name, dep.URL)
		if err != nil {
			return err
		}
	}

	helmArgs := []string{
		"package",
		"-u",
		"--destination", path.Join(dist(), "charts", "addons"),
	}
	err = toolbox().Run(ctx, "helm", append(helmArgs, charts...)...)
	if err != nil {
		return err
	}

	// TODO(dio): Modify timestamp, using the last commit date.
	return nil
}

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
	name := "istio-" + version + "-tetrate-v" + tetrateV
	targz := name + "-linux-amd64.tar.gz"
	downloads := "downloads"
	downloaded := path.Join(downloads, name+".tar.gz")
	build := path.Join("charts", "istio", version)
	dist := path.Join(dist(), "charts", "istio", version)

	_ = os.MkdirAll(downloads, os.ModePerm)
	err := sh.Run("curl", "-sSLo", downloaded, "https://dl.getistio.io/public/raw/files/"+targz)
	if err != nil {
		return err
	}

	err = sh.Run("tar", "xzf", downloaded, "-C", downloads)
	if err != nil {
		return err
	}

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
	_ = os.RemoveAll(dist)
	helmArgs := []string{
		"package",
		"-u",
		"--destination", path.Join(dist),
		"--version", version,
	}
	err = toolbox().Run(ctx, "helm", append(helmArgs, charts...)...)
	if err != nil {
		return err
	}

	// Note: This requires gnu-tar (tar (GNU tar) 1.35).
	// On macOS, brew install gnu-tar.
	infoTime, err := getCreationTime(downloaded)
	if err != nil {
		return err
	}
	return modifyTimestamp(dist, infoTime)
}

func getCreationTime(downloaded string) (string, error) {
	info, err := sh.Output("tar", "--full-time", "-tvf", downloaded)
	if err != nil {
		return "", err
	}
	infoLines := strings.SplitN(info, "\n", 2)
	infoFields := strings.Fields(infoLines[0])
	return formatTime(infoFields[3], infoFields[4]), nil
}

func formatTime(d, t string) string {
	d = strings.ReplaceAll(d, "-", "")
	s := strings.SplitN(t, ":", 3)
	return d + s[0] + s[1] + "." + s[2]
}

func modifyTimestamp(dist, timestamp string) error {
	tgzs, err := os.ReadDir(dist)
	if err != nil {
		return err
	}

	for _, tgz := range tgzs {
		tgzFile := path.Join(dist, tgz.Name())
		err = sh.RunV("touch", "-t", timestamp, tgzFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func getChartVersion(name string) (string, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}

	var c chart
	if err = yaml.Unmarshal(b, &c); err != nil {
		return "", err
	}
	return c.Version, nil
}

type istio struct {
	Version string `yaml:"version"`
	Tetrate string `yaml:"tetrate"`
}

type dependency struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type config struct {
	Istios       []istio      `yaml:"istios"`
	Addons       []string     `yaml:"addons"`
	Dependencies []dependency `yaml:"dependencies"`
}

type chart struct {
	Version string `yaml:"version"`
}

func dist() string {
	return envOr("DIST", "dist")
}

func envOr(key, fallback string) string {
	v := os.Getenv(key)
	if len(v) == 0 {
		return fallback
	}
	return v
}
