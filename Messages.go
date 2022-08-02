package main

import (
	"net"
	"strconv"
	"time"
)

type message struct {
	subnet             net.IPNet // length of array represents subnet mask
	subnetAsBits       []uint8
	finalDestinationAS uint32   // the last AS in the AS path = the final destination = the AS that is responsible for the subnet = the origin of the update message
	peerID             string   // uint16   // represents which of the neighboring peers issued the message //todo change into uint? Then strconv.Itoa(int()) for conversion needed
	timestamp          uint32   // Using unix timestamps.  alternative: time.Time
	aspath             []uint32 // AS path, as written in the AS-path field of the BGP message, only relevant if announcement
	alreadyAnnounced   bool
	/*
		some old peers do not support 4 byte long AS numbers. To be able to communicate with them, the well known AS number
		23456 was introduced. When this AS number appears in an AS path, there is also the AS4path attribute in such a BGP message. In this attribute
		we have the "real" AS path (with 4 byte long AS numbers) stored.

	*/
	isAnnouncement bool // false => message is a withdrawl
	//bgpsubtype     uint16
}

func (m message) toStringNewlines() string {
	result := ""

	if m.isAnnouncement {
		result = result + "ANNOUNCEMENT  for subnet " + m.subnet.String() + "\n"
	} else {
		result = result + "WITHDRAWL     for subnet " + m.subnet.String() + "\n"
	}
	result = result + " issued at " + time.Unix(int64(m.timestamp), 0).String() + " by neighboring peer with ID " + m.peerID + "\n"

	if m.isAnnouncement {
		result = result + "  The final destination AS: " + strconv.Itoa(int(m.finalDestinationAS))
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
	result = result + " issued at " + strconv.Itoa(int(m.timestamp)) + " by neighboring peer (IP) " + m.peerID + ". " //time.Unix(int64(m.timestamp), 0).String()  for readable timestamps

	if m.isAnnouncement {
		result = result + " Final destination AS: " + strconv.Itoa(int(m.finalDestinationAS))
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
