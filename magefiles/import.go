package main

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"strings"

	"github.com/dio/magex/tool"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"gopkg.in/yaml.v3"

	"github.com/tis-system/charts/pkg/helm/repo"
)

var Aliases = map[string]interface{}{
	"import": Import.All,
}

var box *tool.Box

func toolbox() *tool.Box {
	if box == nil {
		box = tool.MustLoadDefault()
	}
	return box
}

type Import mg.Namespace

func (Import) All(ctx context.Context) error {
	_ = os.Setenv("TZ", "UTC")

	err := toolbox().InstallAll(ctx)
	if err != nil {
		return err
	}

	mg.SerialCtxDeps(ctx, Import.Istio, Import.Addons)

	return createIndex()
}

func (Import) Index() error {
	return createIndex()
}

func (Import) Istio(ctx context.Context) error {
	loaded, err := loadConfig()
	if err != nil {
		return err
	}
	return importIstioCharts(ctx, loaded.Istios)
}

func (Import) Addons(ctx context.Context) error {
	loaded, err := loadConfig()
	if err != nil {
		return err
	}

	err = importAddonsCharts(ctx, loaded.Addons, loaded.Dependencies)
	if err != nil {
		return err
	}

	return importSystemAgent(ctx)
}

func containsAnnotation(annotations map[string]string, key string) bool {
	v, ok := annotations[key]
	return ok && v == "true"
}

func createIndex() error {
	indexFile, err := repo.IndexDirectory(dist(), "")
	if err != nil {
		return err
	}

	idx := index{
		Istios:  make(map[string]interface{}, 0),
		Systems: make(map[string]interface{}, 0),
		Addons:  make(map[string]interface{}, 0),
		Demos:   make(map[string]interface{}, 0),
	}
	for k, v := range indexFile.Entries {
		for _, chart := range v {
			if containsAnnotation(chart.Annotations, "tetrate.io/system") {
				idx.Systems[k] = chart
			} else if containsAnnotation(chart.Annotations, "tetrate.io/addon") {
				idx.Addons[k] = chart
			} else if containsAnnotation(chart.Annotations, "tetrate.io/demo") {
				idx.Demos[k] = chart
			} else { // Seems like istio in chart.Keywords is not reliable enough.
				idx.Istios[k] = chart
			}
		}
	}

	b, err := json.Marshal(indexFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &idx.Index)
	if err != nil {
		return err
	}

	b, err = json.Marshal(idx)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(dist(), "index.json"), b, os.ModePerm)
	if err != nil {
		return err
	}

	return indexFile.WriteFile(path.Join(dist(), "index.yaml"), os.ModePerm)
}

type index struct {
	Index   json.RawMessage        `json:"index"`
	Istios  map[string]interface{} `json:"istios"`
	Addons  map[string]interface{} `json:"addons"`
	Demos   map[string]interface{} `json:"demos"`
	Systems map[string]interface{} `json:"systems"`
}

func loadConfig() (*config, error) {
	b, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	loaded := new(config)
	err = yaml.Unmarshal(b, loaded)
	if err != nil {
		return nil, err
	}
	return loaded, nil
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
