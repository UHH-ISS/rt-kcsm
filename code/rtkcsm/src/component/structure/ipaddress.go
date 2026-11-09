package structure

import (
	"net"
	"strconv"
	"strings"
)

const IP_ADDRESS_LENGTH = 17

type IPAddress [IP_ADDRESS_LENGTH]byte

const IS_UNSPECIFIED = 0b0000_0100
const IS_IPV6 = 0b0000_0010
const IS_INTERNAL = 0b0000_0001

func ParseIPAddress(address string) IPAddress {
	ipBytes := net.ParseIP(address)

	ipAddress := IPAddress{}

	if ipBytes != nil {
		if ipBytes.To4() != nil {
			copy(ipAddress[:4], ipBytes.To4()[:4])
		} else {
			copy(ipAddress[:16], ipBytes[:16])
			ipAddress[16] |= IS_IPV6
		}

		if ipBytes.IsUnspecified() {
			ipAddress[16] |= IS_UNSPECIFIED
		}
	} else {
		ipAddress[16] |= IS_UNSPECIFIED
	}

	return ipAddress
}

func (ipAddress IPAddress) IsInternal() bool {
	return ipAddress[16]&IS_INTERNAL == IS_INTERNAL
}

func (ipAddress *IPAddress) SetInternal(isInternal bool) {
	if isInternal {
		ipAddress[16] |= IS_INTERNAL
	} else {
		ipAddress[16] &^= IS_INTERNAL
	}
}

func (ipAddress IPAddress) IsUnspecified() bool {
	return ipAddress[16]&IS_UNSPECIFIED == IS_UNSPECIFIED
}

func (ipAddress *IPAddress) Equal(otherIpAdress *IPAddress) bool {
	return *ipAddress == *otherIpAdress
}

func (ipAddress IPAddress) String() string {
	length := 4
	start := 0
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

func (ipAddress IPAddress) Bytes() [IP_ADDRESS_LENGTH]byte {
	return ipAddress
}
