package values_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tis-system/charts/pkg/helm/values"
)

func TestParse(t *testing.T) {
	p, err := values.FromMarkdown("testdata/a.md")
	require.NoError(t, err)
	_, err = json.Marshal(p)
	require.NoError(t, err)
}
