package main

import (
	"fmt"
	"strconv"
)

type ipv4trieRoot struct {
	childZero *ipv4trie
	childOne  *ipv4trie
}

type trie interface {
	insertMessage(m message, currentDepth uint8) *trie // returns pointer to the node where we ended up inserting the message
	//findConflictsAboveAndSameLevel(c conflicts) conflicts
	//findConflictsBelow(c conflicts) conflicts

	toStringNode() string
	toStringNodeAndSubtrie() string
}

func (ipv4tr *ipv4trieRoot) insert(m message) *trie {

	if m.subnetAsBits[0] == 0 {
		return ipv4tr.childZero.insertMessage(m, 1)
	} else {
		return ipv4tr.childOne.insertMessage(m, 1)
	}
}

type ipv4trie struct { //represents one subnet in the IPv4 range
	childZero          *ipv4trie //0 bit
	childOne           *ipv4trie //1 bit
	parent             *ipv4trie
	representedNet     []uint8
	value              uint8 //0 or 1
	activeAnnouncments []message
}

func (prefixTrie *ipv4trie) insertMessage(m message, currentDepth uint8) *trie {
	subnetMaskLength, _ := m.subnet.Mask.Size()
	var n trie = prefixTrie
	if int(currentDepth) == subnetMaskLength { // right position in trie

		for i := 0; i < len(prefixTrie.activeAnnouncments); i++ {
			if prefixTrie.activeAnnouncments[i].peerID == m.peerID {
				prefixTrie.activeAnnouncments = append(prefixTrie.activeAnnouncments[:i], prefixTrie.activeAnnouncments[i+1:]...) //if there was an announcement before from same peer and with another final destination, we update the message
			}
		}
		if m.isAnnouncement {
			prefixTrie.activeAnnouncments = append(prefixTrie.activeAnnouncments, m)
		}
		/*
				A few words about the logic here in the above 8 lines:
			    in prefixTrie.announcements all announcements for exactly this subnet are stored. If there was a previous announcement from the same peer as our new message that we are currently inserting,
				then there are two possibilities:
				a) the new message is a Withdrawal message. Then the old Announcment by the same peer (and for the same subnet) gets nullified => we delete the old announcement by this peer
				b) the new message is also un Announcement message. Then the old Announcement can be deleted and the new one (with a newer timestamp) can be stored instead
		*/

		return &n
	} else {
		if currentDepth == 32 {
			fmt.Println(Red("m.subnet.Mask.Size() let to: ", subnetMaskLength, ". CurrentDepth is 32. This is the problematic message: \n ", m.toString()))
			return &n
		}

		nextBit := uint8(0)
		nextChild := prefixTrie.childZero
		if m.subnetAsBits[currentDepth] != 0 {
			nextBit = 1
			nextChild = prefixTrie.childOne
		}

		if nextChild == nil {
			reprN := make([]uint8, len(prefixTrie.representedNet)+1)
			copy(reprN, prefixTrie.representedNet)
			reprN[len(prefixTrie.representedNet)] = nextBit
			nextChild = &ipv4trie{parent: prefixTrie, value: nextBit, representedNet: reprN}
			if nextBit == 0 {
				prefixTrie.childZero = nextChild
			} else {
				prefixTrie.childOne = nextChild
			}
		}
		return nextChild.insertMessage(m, currentDepth+1)

	}
}

func (prefixTrie *ipv4trie) toStringNode() string {
	result := ""
	result = result + arrayToString(prefixTrie.representedNet, ", ")

	result = result + "\n Value of Trie = " + strconv.Itoa(int(prefixTrie.value)) + " [For testing: must equal = " + strconv.Itoa(int(prefixTrie.representedNet[len(prefixTrie.representedNet)-1])) + ")"

	if len(prefixTrie.activeAnnouncments) != 0 {
		result = result + "\n    , stored active announcements [" + strconv.Itoa(len(prefixTrie.activeAnnouncments)) + " ]:"
		for _, m := range prefixTrie.activeAnnouncments {
			result = result + "\n       " + m.toString()
		}
	}
	return result
}

func (prefixTrie *ipv4trie) toStringNodeAndSubtrie() string {
	result := "\n"
	result = result + "\n" + "  " + prefixTrie.toStringNode()
	if prefixTrie.childZero != nil {
		result = result + " \n ChildZero " + prefixTrie.childZero.toStringNodeAndSubtrie()
	}
	if prefixTrie.childOne != nil {
		result = result + "\n ChildOne " + prefixTrie.childOne.toStringNodeAndSubtrie()
	}
	return result
}
