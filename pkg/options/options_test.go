package options

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseServers(t *testing.T) {
	testCases := []struct {
		serversStr string
		expect     []*ServerConfig
	}{
		{
			serversStr: "223.5.5.5:53,192.168.34.248:1153",
			expect: []*ServerConfig{
				{
					IP: "223.5.5.5",
					Port: 53,
				},
				{
					IP: "192.168.34.248",
					Port: 1153,
				},
			},
		},
	}

	for _, testCase := range testCases {
		for i, s := range parseServers(testCase.serversStr){
			assert.Equal(t, testCase.expect[i].IP, s.IP)
			assert.Equal(t, testCase.expect[i].Port, s.Port)
		}

	}
}
