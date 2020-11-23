package content

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTrimAssetName(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "root folder",
			input:    "857908472-linuxq3ademo-1.11-6.x86.gz.sh",
			expected: "linuxq3ademo-1.11-6.x86.gz.sh",
		},
		{
			name:     "pak file",
			input:    "baseq3/2483777038-pak0.pk3",
			expected: "baseq3/pak0.pk3",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := trimAssetName(c.input)
			if diff := cmp.Diff(c.expected, result); diff != "" {
				t.Errorf("content: after trimAssetName differs: (-want +got)\n%s", diff)
			}
		})
	}
}
