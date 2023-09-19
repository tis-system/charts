package main

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dio/magex/tool"
	"github.com/magefile/mage/sh"
	"github.com/tis-system/charts/pkg/helm/values"
)

func importAddonsCharts(ctx context.Context, names []string, deps []dependency) error {
	original := path.Join("downloads", "tetratelabs", "helm-charts")
	_ = os.RemoveAll(original)
	_ = os.MkdirAll(original, os.ModePerm)
	err := sh.Run("git", "clone", "--depth", "1", "--single-branch", "-b", "main", "https://github.com/tetratelabs/helm-charts.git", original)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(original)
	if err != nil {
		return err
	}
	err = sh.Run("go", "run", "github.com/3128px/helm-docs/cmd/helm-docs@cd6d19df68d4b5edaf2cd0fcf7a5e096a1e91878")
	if err != nil {
		return err
	}

	err = os.Chdir(cwd)
	if err != nil {
		return nil
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

	for _, chart := range charts {
		_, err = os.Stat(path.Join(chart, "Chart.lock"))
		if err == nil {
			err = toolbox().RunWith(ctx, tool.RunWithOption{Env: map[string]string{
				"TZ": "UTC",
			}}, "helm", "dependency", "build", chart)
			if err != nil {
				return err
			}
		}

		readme := path.Join(chart, "README.md")
		_, err = os.Stat(readme)
		if err == nil {
			valueDir := path.Join(dist(), strings.TrimPrefix(chart, "charts/"))
			_ = os.MkdirAll(valueDir, os.ModePerm)
			parsed, _ := values.FromMarkdown(readme)
			b, err := json.Marshal(parsed)
			if err != nil {
				return err
			}
			_ = os.WriteFile(path.Join(valueDir, "values.json"), b, os.ModePerm)
			err = sh.Run("cp", "-f", readme, valueDir)
			if err != nil {
				return err
			}
		}
	}

	helmArgs := []string{
		"package",
		"--destination", path.Join(dist(), "addons"),
	}
	err = toolbox().RunWith(ctx, tool.RunWithOption{Env: map[string]string{
		"TZ": "UTC",
	}}, "helm", append(helmArgs, charts...)...)
	if err != nil {
		return err
	}

	// TODO(dio): Modify timestamp, using the last commit date.
	return nil
}
