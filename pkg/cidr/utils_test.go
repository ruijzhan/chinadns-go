package cidr

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Line2CidrV4(t *testing.T) {
	cases := []struct {
		line  string
		cidr  string
		valid bool
	}{
		{"apnic|CN|ipv4|43.254.228.0|1024|20140729|allocated", "43.254.228.0/22", true},
		{"apnic|TW|asn|7532|8|19970322|allocated", "", false},
		{"apnic|AU|ipv6|2401:700::|32|20110606|allocated", "", false},
		{"# statement of the location in which any specific resource may", "", false},
	}

	for _, c := range cases {
		cidr, valid := line2CidrV4(c.line)
		assert.Equal(t, c.cidr, cidr)
		assert.Equal(t, c.valid, valid)
	}
}
