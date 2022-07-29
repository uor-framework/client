package attributes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExists(t *testing.T) {
	attributes := Attributes{
		"kind": map[string]struct{}{
			"jpg": {},
			"txt": {},
		},
		"name": map[string]struct{}{
			"fish.jpg": {},
		},
	}
	require.True(t, attributes.Exists("kind", "jpg"))
	require.False(t, attributes.Exists("kind", "png"))
}

func TestFind(t *testing.T) {
	attributes := Attributes{
		"kind": map[string]struct{}{
			"jpg": {},
			"txt": {},
		},
		"name": map[string]struct{}{
			"fish.jpg": {},
		},
	}
	result := attributes.Find("kind")
	require.Len(t, result, 2)
	require.Contains(t, result, "jpg")
	require.Contains(t, result, "txt")
}

func TestAttributes_String(t *testing.T) {
	expString := `kind=jpg,kind=txt,name=fish.jpg`
	attributes := Attributes{
		"kind": map[string]struct{}{
			"jpg": {},
			"txt": {},
		},
		"name": map[string]struct{}{
			"fish.jpg": {},
		},
	}
	require.Equal(t, expString, attributes.String())
}
