package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		content := `CertFile cert.pem
KeyFile key.pem
`
		r := strings.NewReader(content)
		config, err := ParseConfig(r)
		if err != nil {
			t.Fatal(err)
		}

		expected := Config{
			"CertFile": "cert.pem",
			"KeyFile":  "key.pem",
		}

		assert.Equal(t, expected, config)
	})

	t.Run("invalid config", func(t *testing.T) {
		content := `CertFile
`
		r := strings.NewReader(content)
		_, err := ParseConfig(r)

		expected := ParseError{
			LineNumber: 1,
		}
		assert.Equal(t, expected, err)
	})
}
