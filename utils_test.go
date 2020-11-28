package chinadns

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseServers(t *testing.T) {
	testCases := []struct {
		serversStr string
		expect     []*remoteDNS
	}{
		{
			serversStr: "223.5.5.5,192.168.34.248:1153",
			expect: []*remoteDNS{
				{
					IP:   "223.5.5.5",
					Port: 53,
				},
				{
					IP:   "192.168.34.248",
					Port: 1153,
				},
			},
		},
	}

	for _, testCase := range testCases {
		servers, err := parseServers(testCase.serversStr)
		if err != nil {
			t.Fatal(err)
		}
		for i, s := range servers {
			assert.Equal(t, testCase.expect[i].IP, s.IP)
			assert.Equal(t, testCase.expect[i].Port, s.Port)
		}

	}
}
