package cidr

import (
	"bufio"
	"io"
	"math"
	"strconv"
	"strings"
)

func line2CidrV4(line string) (string, bool) {
	if !strings.HasPrefix(line, "apnic|") {
		return "", false
	}
	tokens := strings.Split(line, "|")
	if tokens[2] != "ipv4" {
		return "", false
	}
	mask, err := strconv.Atoi(tokens[4])
	if err != nil {
		return "", false
	}
	mask = 32 - int(math.Log(float64(mask))/math.Log(2))
	return tokens[3] + "/" + strconv.Itoa(mask), true
}

func CirdsByCountry(r io.Reader, cCode string) (cidrs []string) {

	bReader := bufio.NewReader(r)

	for {
		b, _, err := bReader.ReadLine()
		if err == io.EOF {
			break
		}
		line := string(b)
		if strings.Contains(line, cCode) {
			cidr, found := line2CidrV4(line)
			if found {
				cidrs = append(cidrs, cidr)
			}
		}
	}
	return
}
