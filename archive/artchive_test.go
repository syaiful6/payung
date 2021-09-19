package archive

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/helper"
)

func TestRun(t *testing.T) {
	// with nil Archive
	model := config.ModelConfig{
		Archive: nil,
	}
	err := Run(model)
	assert.NoError(t, err)
}

func TestOptions(t *testing.T) {
	includes := []string{
		"/foo/bar/dar",
		"/bar/foo",
		"/ddd",
	}

	excludes := []string{
		"/hello/world",
		"/cc/111",
	}

	opts := options(excludes, includes)
	cmd := strings.Join(opts, " ")
	if helper.IsGnuTar {
		assert.Equal(t, cmd, "--ignore-failed-read -cPf - --exclude=/hello/world --exclude=/cc/111 /foo/bar/dar /bar/foo /ddd")
	} else {
		assert.Equal(t, cmd, "-cPf - --exclude=/hello/world --exclude=/cc/111 /foo/bar/dar /bar/foo /ddd")
	}
}
