package main

import (
	"fmt"
	"github.com/osrg/gobgp/pkg/packet/mrt"
	"math"
	"strconv"
)

var peers []*mrt.Peer //DO NOT USE FOR ANYTHING ELSE BESIDES INITIALIZING "ourPeers". The peers (especially their IP addresses do not stay constant during initialization of trie.
var ourPeers []peer   //stores the peers
var highestPeerId uint32
var highestPeerId100 uint32

type peer struct {
	id uint16
	ip string //this used to be of type net.IP. This led to a bug as sometimes the IP addresses did change without being intended to do so.
	as uint32
}

func (p peer) toString() string {
	result := ""
	result = result + "Peer " + strconv.Itoa(int(p.id)) + ": " + p.ip + " (AS " + strconv.Itoa(int(p.as)) + ")\n"
	return result
}
func ourPeersToString() string {
	result := ""
	for i := 0; i < len(ourPeers); i++ {
		result = result + ourPeers[i].toString()
	}
	return result
}

var peermapByID map[uint16]*peer
var peermapByIP map[string]*peer

func insertPeers() {
	ourPeers = make([]peer, len(peers), int(math.Max(2*float64(len(peers)), 100))) //leave some free capacity in case new peers appear during the updates

	for i := 0; i < len(peers); i++ {
		p := peer{
			id: uint16(highestPeerId),
			ip: peers[i].IpAddress.String(),
			as: peers[i].AS,
		}
		if highestPeerId == 100*highestPeerId100 {
			highestPeerId100++
			fmt.Println(Yellow("peers added: ", highestPeerId))
		}
		if verbose {
			fmt.Println(Yellow("new peer added: ", p.toString()))
		}
		highestPeerId++
		peermapByID[p.id] = &p
		peermapByIP[p.ip] = &p
		ourPeers[i] = p
	}

}

func findPeerIDbyIP(ip string, as uint32) uint16 {
	v, ok := peermapByIP[ip]
	if ok {
		if v.as != as {
			fmt.Println(Red("Found already existing peer with same ip, but for different as"))
		}
		return v.id
	} else {
		p := peer{
			id: uint16(highestPeerId),
			ip: ip,
			as: as,
		}
		if verbose {
			fmt.Println(Yellow("new peer added: ", p.toString()))
		}
		if highestPeerId == highestPeerId100*100 {
			highestPeerId100++
			fmt.Println(Yellow("peers added: ", highestPeerId))
		}
		highestPeerId++
		peermapByID[p.id] = &p
		peermapByIP[p.ip] = &p
		ourPeers = append(ourPeers, p)
		return p.id
	}
}
