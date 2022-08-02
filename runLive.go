/*
The source code in this file is oriented at https://github.com/morrowc/rislive/blob/master/rislive.go .
A copy of their license can be found here: https://github.com/morrowc/rislive/blob/master/LICENSE
Prominent NOTICE: the following code parts have been changed compared to the aforementioned repository
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strconv"
)

// risLive is a struct to hold basic data used in connecting to the RIS Live service and managing data output/collection for the calling client.
type risLive struct {
	url     string
	ua      string
	records int64
	c       chan RisMessage
}

// RisMessage is a single ris_message json message from the ris firehose.
type RisMessage struct {
	Type string          `json:"type"`
	Data *risMessageData `json:"data"`
}

// RisMessageData is the BGP oriented content of the single RisMessage message type.
type risMessageData struct {
	Timestamp float64 `json:"timestamp"`
	Peer      string  `json:"peer"`
	PeerASN   string  `json:"peer_asn,omitempty"`
	ID        string  `json:"id"`
	//Host          string        `json:"host"`
	Type         string        `json:"type"`
	Path         []interface{} `json:"path"`
	DigestedPath []uint32
	//Community     [][]int32          `json:"community"`
	Origin      string   `json:"origin"`
	Withdrawals []string `json:"withdrawals"`

	Announcements []*risAnnouncement `json:"announcements"`

	Raw string `json:"raw"`
}

func (r *risMessageData) toString() string {
	result := ""
	result = result + "timestamp: " + strconv.Itoa(int(r.Timestamp)) + ", peer: " + r.Peer + " (AS = " + r.PeerASN + "), id: " + r.ID + ", type:" + r.Type + "\n"
	result = result + "path: ["
	for i := 0; i < len(r.DigestedPath); i++ {
		result = result + strconv.Itoa(int(r.DigestedPath[i])) + " "
	}
	result = result + "] \n"
	result = result + "origin: " + r.Origin + "\n"
	result = result + "announcements: "
	for i := 0; i < len(r.Announcements); i++ {
		helpResult := ""
		for j := 0; j < len(r.Announcements[i].Prefixes); j++ {
			helpResult = helpResult + r.Announcements[i].Prefixes[j] + " "
		}
		result = result + helpResult + ", "
	}
	result = result + "\nwithdrawls: "
	for i := 0; i < len(r.Withdrawals); i++ {
		result = result + r.Withdrawals[i] + " "
	}
	result = result + "\n \n"
	return result
}

type risAnnouncement struct {
	NextHop  string   `json:"next_hop"`
	Prefixes []string `json:"prefixes"`
}

// NewRisLive creates a new RisLive struct.
func NewRisLive(url, ua string, buffer int) *risLive {
	return &risLive{
		url:     url,
		ua:      ua,
		records: 0,
		c:       make(chan (RisMessage), buffer),
	}
}

func digestPath(m *risMessageData) error {
	m.DigestedPath = []uint32{}
	for _, p := range m.Path {
		var o uint32
		switch v := p.(type) {
		case int:
			o = uint32(v)
		case float64:
			o = uint32(v)
		case []interface{}:
			// Convert p to a slice of interface.
			listSlice, ok := p.([]interface{})
			if !ok {
				return fmt.Errorf("failed to cast path element: %v as %v", p, reflect.TypeOf(p))
			}
			for _, e := range listSlice {
				m.DigestedPath = append(m.DigestedPath, uint32(e.(float64)))
			}
			continue
		default:
			return fmt.Errorf("failed to decode path element: %v as %v", p, reflect.TypeOf(p))
		}
		m.DigestedPath = append(m.DigestedPath, o)
	}
	return nil
}

// Listen connects to the RisLive service, parses the stream into structs
// and makes the data stream available for analysis through the RisLive.Chan channel.
func (r *risLive) Listen() {
restart:
	for {
		var body io.ReadCloser

		fmt.Println("Reading from the firehose...")
		client := &http.Client{}
		req, err := http.NewRequest("GET", r.url, nil)
		if err != nil {
			fmt.Println(Red("failed to create new request to ris-live: %v\n", err))
		}
		//req.Header.Set("User-Agent", r.ua)
		filter := "{\"type\": \"UPDATE\""
		filter = filter + "}"
		fmt.Println(filter)

		req.Header.Add("X-RIS-Subscribe", filter) //Some times the RIPE RIS livestream has problems when answering if there is this header present
		//todo check if needed

		fmt.Println("xris: ", req.Header.Get("X-RIS-Subscribe"))

		//
		fmt.Println(req.Header)
		resp, err := client.Do(req)

		if err != nil {
			fmt.Println(Red("failed to open the http client for action: %v", err))
			return
		}
		fmt.Println(resp.Header)
		fmt.Println(resp)
		defer resp.Body.Close()
		body = resp.Body

		dec := json.NewDecoder(body)
		for {
			var rm RisMessage
			err := dec.Decode(&rm)
			switch {
			case err != nil && err != io.EOF:
				fmt.Println(err)
				fmt.Println(Red("bad json content: \n", rm.Data))

				continue restart
			case err == io.EOF:
				close(r.c)
				return
			}
			err = digestPath(rm.Data)
			if err != nil {
				fmt.Printf(Red("decoding the message data path(%v) failed: %v\n", rm.Data.Path, err))
			}
			r.records++
			r.c <- rm
		}
	}
}

// Get collects messages from the RisLive.Chan channel
func (r *risLive) Get() *risMessageData {
	for rm := range r.c {

		rmd := rm.Data

		if rmd.Type == "UPDATE" {
			return rmd

		}
	}
	return nil
}

func handle(r *risMessageData) {
	//fmt.Println(r.toString())

	var m message
	m.timestamp = uint32(r.Timestamp)
	m.peerID = r.Peer

	complLength := len(r.Withdrawals)
	if len(r.Announcements) > 0 {
		complLength = complLength + len(r.Announcements[0].Prefixes)
	}
	for i := 0; i < complLength; i++ {
		var ipnet *net.IPNet
		var err error
		if i < len(r.Withdrawals) {
			m.isAnnouncement = false
			_, ipnet, err = net.ParseCIDR(r.Withdrawals[i])
		} else {
			m.isAnnouncement = true
			_, ipnet, err = net.ParseCIDR(r.Announcements[0].Prefixes[i-len(r.Withdrawals)])
			m.aspath = r.DigestedPath
			if len(r.DigestedPath) > 0 {
				m.finalDestinationAS = r.DigestedPath[len(r.DigestedPath)-1]
			} else {
				fmt.Println(Red("Digested Path length was 0!"))
				fmt.Println(r.toString())
				return
			}
		}
		if err != nil {
			fmt.Println(Red("Could not parse to subnet", err))
		}
		m.subnet = *ipnet
		if m.subnet.IP.To4() != nil {
			m.subnetAsBits = convertIPtoBits(m.subnet)

			fmt.Println(m.toString())
			fmt.Println("______________________________________________")
		}
	}
}

func runInLiveMode() {
	r := NewRisLive(liveStream, risClient, buffer)
	go r.Listen()

	for true {
		result := r.Get()
		if len(result.Announcements) > 1 {
			if result.Announcements[0].Prefixes[0] != result.Announcements[1].Prefixes[0] {
				fmt.Println(Red("we have received a message with multiple announcements (regarding json format) with two different values as first prefix: \n", result.toString()))
			}
		}
		//fmt.Println("received this message: \n", result.toString())
		handle(result)
	}
}
