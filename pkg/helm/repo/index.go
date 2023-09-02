package repo

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/provenance"
	"helm.sh/helm/v3/pkg/repo"
)

func IndexDirectory(dir, baseURL string) (*repo.IndexFile, error) {
	archives, err := filepath.Glob(filepath.Join(dir, "*/*.tgz"))
	if err != nil {
		return nil, err
	}
	moreArchives, err := filepath.Glob(filepath.Join(dir, "*/*/*.tgz"))
	if err != nil {
		return nil, err
	}
	archives = append(archives, moreArchives...)

	index := repo.NewIndexFile()
	for _, arch := range archives {
		fname, err := filepath.Rel(dir, arch)
		if err != nil {
			return index, err
		}

		var parentDir string
		parentDir, fname = filepath.Split(fname)
		// filepath.Split appends an extra slash to the end of parentDir. We want to strip that out.
		parentDir = strings.TrimSuffix(parentDir, string(os.PathSeparator))
		parentURL, err := urlJoin(baseURL, parentDir)
		if err != nil {
			parentURL = path.Join(baseURL, parentDir)
		}

		c, err := loader.Load(arch)
		if err != nil {
			// Assume this is not a chart.
			continue
		}
		hash, err := provenance.DigestFile(arch)
		if err != nil {
			return index, err
		}
		if err := mustAdd(index, c.Metadata, fname, arch, parentURL, hash); err != nil {
			return index, fmt.Errorf("failed adding to %s to index: %w", fname, err)
		}
	}
	return index, nil
}

func urlJoin(baseURL string, paths ...string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	// We want path instead of filepath because path always uses /.
	all := []string{u.Path}
	all = append(all, paths...)
	u.Path = path.Join(all...)
	return u.String(), nil
}

func mustAdd(i *repo.IndexFile, md *chart.Metadata, filename, file, baseURL, digest string) error {
	if i.Entries == nil {
		return errors.New("entries not initialized")
	}

	if md.APIVersion == "" {
		md.APIVersion = chart.APIVersionV1
	}
	if err := md.Validate(); err != nil {
		return fmt.Errorf("validate failed for %s: %w", filename, err)
	}

	u := filename
	if baseURL != "" {
		_, file := filepath.Split(filename)
		var err error
		u, err = urlJoin(baseURL, file)
		if err != nil {
			u = path.Join(baseURL, file)
		}
	}

	info, err := os.Stat(file)
	if err != nil {
		return err
	}

	cr := &repo.ChartVersion{
		URLs:     []string{u},
		Metadata: md,
		Digest:   digest,
		Created:  info.ModTime().UTC(),
	}
	ee := i.Entries[md.Name]
	i.Entries[md.Name] = append(ee, cr)
	return nil
}
