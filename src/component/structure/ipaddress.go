package structure

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type IPAddress [17]byte

const IS_UNSPECIFIED = 0b0000_0100
const IS_IPV6 = 0b0000_0010
const IS_PRIVATE = 0b0000_0001

func ParseIPAddress(address string) IPAddress {
	ipBytes := net.ParseIP(address)

	ipAddress := IPAddress{}

	if ipBytes != nil {
		ipAddress[0] = ipBytes[0]
		ipAddress[1] = ipBytes[1]
		ipAddress[2] = ipBytes[2]
		ipAddress[3] = ipBytes[3]
		ipAddress[4] = ipBytes[4]
		ipAddress[5] = ipBytes[5]
		ipAddress[6] = ipBytes[6]
		ipAddress[7] = ipBytes[7]
		ipAddress[8] = ipBytes[8]
		ipAddress[9] = ipBytes[9]
		ipAddress[10] = ipBytes[10]
		ipAddress[11] = ipBytes[11]
		ipAddress[12] = ipBytes[12]
		ipAddress[13] = ipBytes[13]
		ipAddress[13] = ipBytes[13]
		ipAddress[14] = ipBytes[14]
		ipAddress[15] = ipBytes[15]

		if ipBytes.IsPrivate() {
			ipAddress[16] |= IS_PRIVATE
		}

		if ipBytes.To4() == nil {
			ipAddress[16] |= IS_IPV6
		}

		if ipBytes.IsUnspecified() {
			ipAddress[16] |= IS_UNSPECIFIED
		}
	}

	return ipAddress
}

func (ipAddress IPAddress) IsPrivate() bool {
	return ipAddress[16]&IS_PRIVATE == IS_PRIVATE
}

func (ipAddress IPAddress) IsUnspecified() bool {
	return ipAddress[16]&IS_UNSPECIFIED == IS_UNSPECIFIED
}

func (ipAddress IPAddress) Equal(otherIpAdress IPAddress) bool {
	return ipAddress == otherIpAdress
}

func (ipAddress IPAddress) IsSameSubnet(otherIpAdress IPAddress) bool {
	networkSize := 24
	if ipAddress[16]&IS_IPV6 == IS_IPV6 {
		networkSize = 64
	}

	otherNetworkSize := 24
	if otherIpAdress[16]&IS_IPV6 == IS_IPV6 {
		otherNetworkSize = 64
	}

	if networkSize != otherNetworkSize {
		return false
	}

	_, network, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ipAddress.String(), networkSize))
	if err != nil {
		return false
	}

	return network.Contains(net.ParseIP(otherIpAdress.String()))
}

func (ipAddress IPAddress) String() string {
	length := 4
	start := 12
	separator := "."
	base := 10

	if ipAddress[16]&IS_IPV6 == IS_IPV6 {
		separator = ":"
		start = 0
		length = 16
		base = 16
	}

	str := []string{}

	for i := range length {
		str = append(str, strconv.FormatInt(int64(ipAddress[start+i]), base))
	}

	return strings.Join(str, separator)
}
