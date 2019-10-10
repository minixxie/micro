package micro

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitJaeger(t *testing.T) {
	closer, err := InitJaeger("", "localhost:6831", "localhost:6831", true)
	assert.Nil(t, closer)
	assert.Error(t, err)
}
