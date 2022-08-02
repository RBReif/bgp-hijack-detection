package main

import "net"

func convertIPtoBits(ipNet net.IPNet) []uint8 {
	var ip = ipNet.IP
	var asField []uint8
	length, _ := ipNet.Mask.Size()
	if ip.To4() != nil {
		for i := 0; i < 4; i++ {
			x := int(ip.To4()[i])
			for j := 7; j >= 0; j-- {
				if x >= 1<<j {
					asField = append(asField, 1)
					x = x - (1 << j)
				} else {
					asField = append(asField, 0)
				}
			}
		}
	} else {
		for i := 0; i < 16; i++ {

			x := int(ip.To16()[i])
			for j := 7; j >= 0; j-- {
				if x >= 1<<j {
					asField = append(asField, 1)
					x = x - (1 << j)

				} else {
					asField = append(asField, 0)
				}
			}
		}

	}
	if length > cap(asField) {
		length = cap(asField)
	}

	return asField[0:length]
}
