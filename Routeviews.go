package main

import (
	"fmt"
	"github.com/osrg/gobgp/pkg/packet/bgp"
	"net"

	"github.com/osrg/gobgp/pkg/packet/mrt"
	"log"
	"time"
)

var bgpd BGPDump

type BGPDump struct {
	Date time.Time
}

func (b *BGPDump) parseRIBAndInsert(ribFileName string) error {
	scanner, err := getRightScanner(ribFileName)
	if err != nil {
		return err
	}
	scanner.Split(mrt.SplitMrt)

	countSplitsInRIB := 0
	countMrtRibs := 0
	countSingleEntries := 0
	indexTableCount := 0
entries:
	for scanner.Scan() {
		countSplitsInRIB++

		if countSplitsInRIB > 10000000 { //for debugging. Vary threshold to preferred amount of read RIB entries
			fmt.Println("Too many entries. STOPPING....")
			break
		}
		var m message

		data := scanner.Bytes()
		hdr := &mrt.MRTHeader{}
		errh := hdr.DecodeFromBytes(data[:mrt.MRT_COMMON_HEADER_LEN])
		if errh != nil { //changed to errh (before there stood err (probably a mistake)
			return errh
		}
		msg, err := mrt.ParseMRTBody(hdr, data[mrt.MRT_COMMON_HEADER_LEN:])
		if err != nil {
			log.Printf("could not parse mrt body: %v", err)
			continue entries
		}
		if msg.Header.Type != mrt.TABLE_DUMPv2 {
			return fmt.Errorf("unexpected message type: %d", msg.Header.Type)
		}

		//m.timestamp = msg.Header.Timestamp
		m.isAnnouncement = true //RIBS are all announcements
		//m.bgpsubtype = msg.Header.SubType

		switch mtrBody := msg.Body.(type) {

		case *mrt.PeerIndexTable:
			indexTableCount++
			peers = msg.Body.(*mrt.PeerIndexTable).Peers //added
			if indexTableCount != 1 {
				return fmt.Errorf("got > 1 PeerIndexTable")
			}
			insertPeers()

		case *mrt.Rib:
			countMrtRibs++
			rib := msg.Body.(*mrt.Rib)
			prefix := rib.Prefix

			_, ipnet, err := net.ParseCIDR(prefix.String())
			if err != nil {
				fmt.Println(Red("Could not parse to subnet", err))
			}
			m.subnet = *ipnet
			if m.subnet.IP.To4() != nil {

				m.subnetAsBits = convertIPtoBits(m.subnet)
			}
			if len(rib.Entries) < 0 {
				return fmt.Errorf("no entries")
			}

			for i := 0; i < len(rib.Entries); i++ {
				countSingleEntries++
				m.timestamp = uint32(rib.Entries[i].OriginatedTime)
				m.peerID = ourPeers[rib.Entries[i].PeerIndex].id //during the initialization PeerIndex should already be equal to id
				setASpathInMessage(&m, rib.Entries[i].PathAttributes)

				insertAndFindConflicts(m, false)
			}

		default:
			return fmt.Errorf("unsupported message %v %s", mtrBody)
		}

	}

	fmt.Println(Teal("items in RIB file, seperated by MRT splitting function: ", countSplitsInRIB))
	fmt.Println(Teal("of these items, how many were of type mrt.Ribs:         ", countMrtRibs))
	fmt.Println(Teal("all together there were how many single RIB entries:    ", countSingleEntries))
	fmt.Println(Teal("found peer index tables:                                 ", indexTableCount))
	return nil
}

var messages BGPDump

func (b *BGPDump) parseUpdatesAndInsert(inputFile string, findConflicts bool) error {
	scanner, err := getRightScanner(inputFile)
	if err != nil {
		return err
	}
	scanner.Split(mrt.SplitMrt)

	//some non-essential logging variables to keep overview
	countEntries := 0
	countNotRelevantMRTBody := 0
	countNotRelevantBGPMsgBody := 0
	countNeitherUpdateNorWithdrawl := 0
	//	countConflictTriggers := 0

entries:
	for scanner.Scan() {
		if countEntries > 100000000 { //for debugging. vary threshold to desired limit
			fmt.Println(Red("Too many entries in updates. STOPPING..."))
			break
		}
		countEntries++
		//extracting the bytes for the next MRT entry
		data := scanner.Bytes()

		//extracting the MRT header from the bytes
		hdr := &mrt.MRTHeader{}
		errh := hdr.DecodeFromBytes(data[:mrt.MRT_COMMON_HEADER_LEN])
		if errh != nil {
			return errh
		}

		/*
			The BGP4MP_ET type is not implemented in the mrt-Module.
			The difference to BGP4MP is only a more precise timestamp (containting microseconds) in an extra field.
			To be able to use the unmodified mrt-Module, we cast a BGP4MP_ET message to a BGP4MP message and "jump over" the
			extra field containing the microseconds. The normal timestamp is of course parsed as normal.
		*/
		var skip uint32
		skip = 0
		if hdr.Type == mrt.BGP4MP_ET {
			skip = 4              //we will later "jump over" the 4 bytes field containing the microseconds
			hdr.Type = mrt.BGP4MP //we change the type indicator from BGP4MP_ET to BGP4MP
		}

		//extracting the MRT Message. The MRT Header is already extracted. The MRT Body is extracted in the following. Both are stored in msg
		var msg *mrt.MRTMessage
		var err error
		hdr.Len = hdr.Len - skip
		msg, err = mrt.ParseMRTBody(hdr, data[mrt.MRT_COMMON_HEADER_LEN+skip:])
		if err != nil {
			log.Printf("could not parse mrt body: %v", err)
			continue entries
		}

		switch msg.Body.(type) {
		case *mrt.BGP4MPMessage: //we expect the type of the MRT Body to be of type BGP4MPMessage
			bgp4msg := msg.Body.(*mrt.BGP4MPMessage)
			//A BGP4MPMessage contains itself a BGPMessage consisting of a Body and a header
			bgpmsg := bgp4msg.BGPMessage

			switch bgpmsg.Body.(type) {
			case *bgp.BGPUpdate: //we expect the body to be of type BGPUpdate
				bgpmsgBody := bgpmsg.Body.(*bgp.BGPUpdate)

				//we start with the creation of a new instance of our Message type with the already extracted attributes
				var m message
				m.timestamp = msg.Header.Timestamp
				//m.bgpsubtype = msg.Header.SubType

				peerip := bgp4msg.PeerIpAddress.String()
				peerAs := bgp4msg.PeerAS

				m.peerID = findPeerIDbyIP(peerip, peerAs)

				//we make sure that we either have a BGP announcement or withdrawal
				if len(bgpmsgBody.NLRI) == 0 && len(bgpmsgBody.WithdrawnRoutes) == 0 {
					countNeitherUpdateNorWithdrawl++
					fmt.Println(Red("BGP Message seems to be neither Announcement nor Withdrawal: ", bgpmsgBody))
					continue entries
				}
				subnetsAnouncments := bgpmsgBody.NLRI           //if it is an announcement we extract the announced prefixes from NLRI and store them in subnets
				subnetsWithdrawls := bgpmsgBody.WithdrawnRoutes //if it is a withdrawal we extract the withdrawn prefixes from WithdrawnRoutes and store them in subnets

				for i := 0; i < len(subnetsAnouncments)+len(subnetsWithdrawls); i++ { //for each announced or withdrawn subnet we create a single instance of type message and insert it into our trie
					var ipnet *net.IPNet
					var err error
					if i < len(subnetsAnouncments) {
						m.isAnnouncement = true
						_, ipnet, err = net.ParseCIDR(subnetsAnouncments[i].String())
					} else {
						m.isAnnouncement = false
						_, ipnet, err = net.ParseCIDR(subnetsWithdrawls[i-len(subnetsAnouncments)].String())
					}
					if err != nil {
						fmt.Println(Red("Could not parse to subnet", err))
					}

					m.subnet = *ipnet
					if m.subnet.IP.To4() != nil {

						m.subnetAsBits = convertIPtoBits(m.subnet)
					}
					//if we have an Announcement we also have to set the AS path (and AS4path) attributes
					if m.isAnnouncement {
						setASpathInMessage(&m, bgpmsgBody.PathAttributes)
					}

					//we insert the current message in the IPv4 or IPv6 trie
					insertAndFindConflicts(m, findConflicts)
				}

			default:
				countNotRelevantBGPMsgBody++
			}
		default:
			countNotRelevantMRTBody++
			fmt.Println(Red("MRT Body is not of type BGP4MPMessage, but of: ", msg.Header.Type))
		}
		//fmt.Println(msg)
		//fmt.Println()
	}
	fmt.Println(Teal("read update entries:                                          ", countEntries))
	fmt.Println(Teal("not parsable MRT body (not of type BGP4MPMessage):            ", countNotRelevantMRTBody))
	fmt.Println(Teal("not parsable BGP message body (not of type BGPUpdate):        ", countNotRelevantBGPMsgBody))
	fmt.Println(Teal("BGP update messages with neither announcements nor withdrawls:", countNeitherUpdateNorWithdrawl))

	return nil
}

func setASpathInMessage(m *message, attributes []bgp.PathAttributeInterface) { //this function sets the aspath  in a message m

	//In case there are multiple AS paths, we want to choose one of the smallest AS paths
	minimumLengthOfBestASPath := 10000
	var bestASPath []uint32
	minimumLengthOfBestRealASPath := 10000
	var bestRealASPath []uint32
attrs:
	// we iterate over the attributes
	for i := 0; i < len(attributes); i++ {
		switch pa := attributes[i].(type) {
		case *bgp.PathAttributeAsPath: // we have an AS path
			if len(pa.Value) < 1 {
				continue attrs
			}
			if len(pa.Value) > 1 {
			}

			if v, ok := pa.Value[0].(*bgp.As4PathParam); ok { // we expect the AS paths to be of kind "AS4PathParam"
				if len(v.AS) < 1 { //reminder: there stood 0. this was probably wrong
					continue attrs
				}
				if len(v.AS) < minimumLengthOfBestASPath { // we compare the path length to our previous best path length
					minimumLengthOfBestASPath = len(v.AS)
					bestASPath = v.AS
				}
			}
			if v, ok := pa.Value[0].(*bgp.AsPathParam); ok { // deprecated but still in use
				if len(v.AS) < 1 { //reminder there stood 0. this was probably wrong
					continue attrs
				}
				if len(v.AS) < minimumLengthOfBestASPath {
					minimumLengthOfBestASPath = len(v.AS)

					//we have to cast each uint16  in the slice into an uint32
					bestASPath = make([]uint32, len(v.AS))
					for i, v2 := range v.AS {
						bestASPath[i] = uint32(v2)
					}
				}
			}

		case *bgp.PathAttributeAs4Path:
			/*
				some old peers do not support 4 byte long AS numbers. To be able to communicate with them, the well known AS number
				23456 was introduced. When this AS number appears in an AS path, we also have the AS4path attribute. In this attribute
				we have the "real" AS path (with 4 byte long AS numbers) stored.
			*/
			if len(pa.Value) < 1 {
				continue attrs
			}
			pv := pa.Value[0]
			if len(pv.AS) < 1 {
				continue attrs
			}
			if len(pv.AS) < minimumLengthOfBestRealASPath {
				minimumLengthOfBestRealASPath = len(pv.AS)
				bestRealASPath = pv.AS //hint: if we get problems here, maybe pv.AS is not of type AS4PathParam but ASPathParam
			}
		default:
			//	fmt.Println(" no PathAttributeAS(4)Path but ", pa)
		}

	}
	m.aspath = bestASPath // we set the aspath attribute in message m
	if bestASPath != nil {
		m.finalDestinationAS = bestASPath[len(bestASPath)-1]
	} else {
		fmt.Println("no best AS path specified. Could not set origin AS")
	}
	if minimumLengthOfBestRealASPath != 10000 { // if needed, we set the aspath to the "real" ASPath
		realAsPath := make([]uint32, len(bestRealASPath))
		copy(realAsPath, bestRealASPath)
		realAsPath = append([]uint32{m.aspath[0]}, realAsPath...) // it seems that the as4path does not include the first AS number. We want to see the full path and prepend this AS number
		m.aspath = realAsPath
		m.finalDestinationAS = realAsPath[len(realAsPath)-1]
	}

}
