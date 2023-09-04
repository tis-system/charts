package values

import (
	"bufio"
	"os"
	"strings"
)

type Parsed struct {
	Values []*Value `json:"values"`
}

type Value struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

func FromMarkdown(name string) (*Parsed, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, nil
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var parsed = &Parsed{
		Values: make([]*Value, 0),
	}

	var found bool
	var first bool
	for scanner.Scan() {
		text := scanner.Text()

		if len(text) == 0 {
			continue
		}

		if text == "|-----|------|---------|-------------|" {
			found = true
			first = true
		}

		if found {
			if first {
				first = false
				continue
			}
			if strings.HasPrefix(text, "|") {
				cols := strings.Split(text, "|")
				value := &Value{
					Key:         strings.TrimSpace(cols[1]),
					Type:        strings.TrimSpace(cols[2]),
					Default:     strings.ReplaceAll(strings.TrimSpace(cols[3]), "`", ""),
					Description: strings.TrimSpace(cols[4]),
				}
				parsed.Values = append(parsed.Values, value)
			}
		}
	}

	return parsed, nil
}
