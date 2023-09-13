package main

import (
	"context"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/dio/magex/tool"
)

func importSystemAgent(ctx context.Context) error {
	base := path.Join("charts", "system")
	entries, err := os.ReadDir(base)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		s := path.Join(base, entry.Name())
		helmArgs := []string{
			"package",
			"-u",
			"--destination", path.Join(dist(), "system"),
		}

		v, err := readVersion(s)
		if err == nil {
			version := strings.TrimPrefix(v, "v")
			helmArgs = append(helmArgs, "--app-version", version, path.Join(s), "--version", version)
		}

		if err := toolbox().RunWith(ctx, tool.RunWithOption{Env: map[string]string{
			"TZ": "UTC",
		}}, "helm", helmArgs...); err != nil {
			return err
		}
	}

	// TODO(dio): Modify timestamp, using the last commit date.
	return nil
}

func readVersion(name string) (string, error) {
	entries, err := os.ReadDir(name)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.Name() == "VERSION" {
			content, err := os.ReadFile(path.Join(name, entry.Name()))
			if err != nil {
				return "", err
			}
			return string(content), nil
		}
	}
	return "", errors.New("failed to get version")
}
