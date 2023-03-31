package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testPath1 = "test"
	testPath2 = "test2"
)

func TestGetPiTemp(t *testing.T) {
	f1, err := os.OpenFile(testPath1, os.O_RDWR|os.O_CREATE, 0755)
	assert.NoError(t, err)
	f1.WriteString("34503")

	f2, err := os.OpenFile(testPath2, os.O_RDWR|os.O_CREATE, 0755)
	assert.NoError(t, err)
	f2.WriteString("3425")

	defer func() {
		os.Remove(testPath1)
		os.Remove(testPath2)
	}()

	temp, err := GetPiTemp(&testPath1)
	assert.NoError(t, err)
	assert.Equal(t, float64(34.5), temp)

	_, err = GetPiTemp(&testPath2)
	assert.Error(t, err)
}
