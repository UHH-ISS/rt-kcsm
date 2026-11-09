package structure

import (
	"net"
)

type IPNetwork IPAddress

func ParseIPNetwork(network string) IPNetwork {
	_, networkStruct, _ := net.ParseCIDR(network)

	ipNetwork := IPNetwork{}
	if networkStruct.IP.To4() != nil {
		copy(ipNetwork[:4], networkStruct.IP.To4())
	} else {
		copy(ipNetwork[:16], networkStruct.IP.To16())
	}

	ones, _ := networkStruct.Mask.Size()
	ipNetwork[16] = byte(ones)

	return ipNetwork
}

func (ipNetwork *IPNetwork) HasIpAddress(ipAddress *IPAddress) bool {
	fullBytes := ipNetwork[16] / 8
	remainingBits := ipNetwork[16] % 8

	for i := range fullBytes {
		if ipAddress[i] != ipNetwork[i] {
			return false
		}
	}

	if remainingBits > 0 {
		mask := byte(0b1111_1111 << (8 - remainingBits))

		if (ipAddress[fullBytes] & mask) != (ipNetwork[fullBytes] & mask) {
			return false
		}
	}

	return true
}

func (ipNetwork *IPNetwork) Equal(otherIPNetwork *IPNetwork) bool {
	return *ipNetwork == *otherIPNetwork
}

type IPNetworks []IPNetwork

func (networks IPNetworks) Len() int {
	return len(networks)
}

func (networks IPNetworks) Less(i int, j int) bool {
	return networks[i][16] < networks[j][16]
}

func (networks IPNetworks) Swap(i int, j int) {
	networks[i], networks[j] = networks[j], networks[i]
}
