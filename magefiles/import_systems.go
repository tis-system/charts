package main

import (
	"context"
	"path"

	"github.com/dio/magex/tool"
)

func importSystemAgent(ctx context.Context) error {
	helmArgs := []string{
		"package",
		"-u",
		"--destination", path.Join(dist(), "system"),
	}
	err := toolbox().RunWith(ctx, tool.RunWithOption{Env: map[string]string{
		"TZ": "UTC",
	}}, "helm", append(helmArgs, path.Join("charts", "system", "agent"))...)
	if err != nil {
		return err
	}

	// TODO(dio): Modify timestamp, using the last commit date.
	return nil
}
