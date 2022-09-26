package main

import (
	"net"
	"strconv"
	"time"
)

type message struct {
	subnet           net.IPNet // length of array represents subnet mask
	subnetAsBits     []uint8
	origin           uint32   // the last AS in the AS path = the final destination = the AS that is responsible for the subnet = the origin of the update message
	peerID           uint16   // represents which of the neighboring peers issued the message
	timestamp        uint32   // Using unix timestamps.  alternative: time.Time
	aspath           []uint32 // AS path, as written in the AS-path field of the BGP message, only relevant if the message is an announcement
	alreadyAnnounced bool     //prevents that the same (still active) conflict is found over and over again by the same update message

	isAnnouncement  bool // false => message is a withdrawal
	isSpecialPrefix bool
}

func (m message) toStringNewlines() string {
	result := ""

	if m.isAnnouncement {
		result = result + "ANNOUNCEMENT  for subnet " + m.subnet.String() + "\n"
	} else {
		result = result + "WITHDRAWAL     for subnet " + m.subnet.String() + "\n"
	}
	result = result + " issued at " + time.Unix(int64(m.timestamp), 0).String() + " by neighboring peer with ID " + strconv.Itoa(int(m.peerID)) + "\n"

	if m.isAnnouncement {
		result = result + "  The origin AS AS: " + strconv.Itoa(int(m.origin))
		result = result + "  The AS path is the following: " + aspathtoString(m.aspath) + "\n"
		result = result + "\n"
	}
	return result
}
func (m message) toString() string {
	result := ""
	if m.isAnnouncement {
		result = result + "A for subnet " + m.subnet.String() + ": "
	} else {
		result = result + "W for subnet " + m.subnet.String() + ": "
	}
	result = result + " issued at " + strconv.Itoa(int(m.timestamp)) + " by neighboring peer (ID) " + strconv.Itoa(int(m.peerID)) + ". " //time.Unix(int64(m.timestamp), 0).String()  for readable timestamps

	if m.isAnnouncement {
		result = result + " Origin AS: " + strconv.Itoa(int(m.origin))
		result = result + "  AS path: " + aspathtoString(m.aspath)
	}
	return result
}

func aspathtoString(aspath []uint32) string {

	aspathAsString := ""
	for i := 0; i < len(aspath); i++ {
		if i != 0 {
			aspathAsString = aspathAsString + " - "
		}
		aspathAsString = aspathAsString + strconv.Itoa(int(aspath[i]))
	}
	return aspathAsString
}
