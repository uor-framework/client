package descriptor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnnotationsToAttributes(t *testing.T) {
	expList := map[string][]string{
		"kind": {"jpg"},
		"name": {"fish.jpg"},
	}
	annotations := map[string]string{
		"kind": "jpg",
		"name": "fish.jpg",
	}
	require.Equal(t, expList, AnnotationsToAttributes(annotations).List())
}
