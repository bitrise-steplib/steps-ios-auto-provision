package certstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToP12(t *testing.T) {
	s, err := New()
	require.NoError(t, err)
	certs, err := s.Download("file:///Users/godrei/Downloads/NewBitfallDevDistrCertificates.p12", "")
	require.NoError(t, err)
	for _, c := range certs {
		pth, err := s.ToP12(c)
		fmt.Println(pth)
		require.NoError(t, err)
	}
	require.Error(t, err)
}
