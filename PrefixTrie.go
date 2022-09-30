package main

import (
	"fmt"
	"strconv"
	"time"
)

type ipv4trieRoot struct {
	childZero *ipv4trie
	childOne  *ipv4trie
}

type trie interface {
	insertMessage(m message, currentDepth uint8) *trie // returns pointer to the node where we ended up inserting the message
	findConflictsBelow(c conflicts) conflicts
	findConflictsAboveAndSameLevel(c conflicts) conflicts
	toStringNode() string
	toStringNodeAndSubtrie() string
	isRelevant() bool
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
	relevant           bool
}

func (prefixTrie *ipv4trie) isRelevant() bool {
	return prefixTrie.relevant
}

func (prefixTrie *ipv4trie) insertMessage(m message, currentDepth uint8) *trie {
	subnetMaskLength, _ := m.subnet.Mask.Size()
	var n trie = prefixTrie

	if int(currentDepth) == subnetMaskLength { // right position in trie
		if m.isSpecialPrefix {
			prefixTrie.relevant = true
			return &n
		}

		for i := 0; i < len(prefixTrie.activeAnnouncments); i++ {

			if m.isAnnouncement {
				if prefixTrie.activeAnnouncments[i].origin == m.origin {
					m.alreadyAnnounced = true // to prevent the same (still active) conflict to be found over and over again
				}
			}

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
			nextChild = &ipv4trie{parent: prefixTrie, value: nextBit, representedNet: reprN, relevant: prefixTrie.relevant}
			if nextBit == 0 {
				prefixTrie.childZero = nextChild
			} else {
				prefixTrie.childOne = nextChild
			}
		}
		return nextChild.insertMessage(m, currentDepth+1)

	}
}

func (prefixTrie *ipv4trie) findConflictsThisLevel(c conflicts) conflicts {
	if prefixTrie.activeAnnouncments != nil {
		for i := len(prefixTrie.activeAnnouncments); i > 0; i-- { //we iterate over all announcements. from newest to oldest
			if c.referenceAnnouncement.origin != prefixTrie.activeAnnouncments[i-1].origin { //it´s not a conflict if the final destination AS is the same
				//we have found a potential conflict

				alreadyAConflictByDifferentPeer := false
				for j := 0; j < len(c.conflictingMessages); j++ {
					if c.conflictingMessages[j].origin == prefixTrie.activeAnnouncments[i-1].origin &&
						c.conflictingMessages[j].subnet.String() == prefixTrie.activeAnnouncments[i-1].subnet.String() {
						alreadyAConflictByDifferentPeer = true
					}
				}
				if !alreadyAConflictByDifferentPeer {
					c.conflictingMessages = append(c.conflictingMessages, prefixTrie.activeAnnouncments[i-1])
					if prefixTrie.relevant {
						c.relevant = true
					}
				}
			}
		}
	}

	return c
}

func (prefixTrie *ipv4trie) findConflictsBelow(c conflicts) conflicts {
	size, _ := c.referenceAnnouncement.subnet.Mask.Size()
	if size == len(prefixTrie.representedNet) { // we are not interested in conflicts at the same level as the reference message (they were already found in findConflictsThislevelAndAbove
		for i := 0; i < len(prefixTrie.activeAnnouncments); i++ {
			if prefixTrie.activeAnnouncments[i].peerID == c.referenceAnnouncement.peerID {
				if prefixTrie.activeAnnouncments[i].alreadyAnnounced {
					return c
				}
			}
		}
	} else {
		c = prefixTrie.findConflictsThisLevel(c)
	}

	if prefixTrie.childZero != nil {
		c = prefixTrie.childZero.findConflictsBelow(c)
	}
	if prefixTrie.childOne != nil {
		c = prefixTrie.childOne.findConflictsBelow(c)
	}
	return c
}

func (prefixTrie *ipv4trie) findConflictsAboveAndSameLevel(c conflicts) conflicts {
	size, _ := c.referenceAnnouncement.subnet.Mask.Size()
	if size == len(prefixTrie.representedNet) {
		for i := 0; i < len(prefixTrie.activeAnnouncments); i++ {
			if prefixTrie.activeAnnouncments[i].peerID == c.referenceAnnouncement.peerID {
				if prefixTrie.activeAnnouncments[i].alreadyAnnounced {
					return c //if we land here, there was already the same announcement from the same peer and we already found all respective conflicts
				}
			}
		}
	}

	c = prefixTrie.findConflictsThisLevel(c)

	if prefixTrie.parent != nil {
		return prefixTrie.parent.findConflictsAboveAndSameLevel(c) //we go recursively up to the root/ to less specific subnets
	} else { //we are at the first bit.
		return c
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

func insertAndFindConflicts(m message, findConflicts bool) {
	if stopT.IsZero() || time.Now().Before(stopT) {
		subnetSize, _ := m.subnet.Mask.Size()
		if m.subnet.IP.To4() != nil && len(m.subnetAsBits) <= 32 && subnetSize <= 32 {
			o, _ := m.subnet.Mask.Size()
			if o > 0 {

				nodeWhereInserted := *ipv4T.insert(m)
				//fmt.Println("going to insert message: \n", m.toString(), "\n")

				if m.isAnnouncement && findConflicts {
					conflictsField := make([]message, 0)

					c := conflicts{
						referenceIPasField:    convertIPtoBits(m.subnet),
						referenceAnnouncement: m,
						conflictingMessages:   conflictsField,
						relevant:              nodeWhereInserted.isRelevant(),
					}
					confl := nodeWhereInserted.findConflictsAboveAndSameLevel(c)
					confl = nodeWhereInserted.findConflictsBelow(confl)
					if len(confl.conflictingMessages) > 0 {
						countConflicts = countConflicts + len(confl.conflictingMessages)
						prepareJSON(confl)
						updateSummary(confl)
						if countConflictTriggers == 1000*countConflictTriggers1000 {
							countConflictTriggers1000++
							fmt.Println(White("Messages that triggered conflicts so far: ", countConflictTriggers))
						}
						countConflictTriggers++

						if verbose {
							fmt.Println(White(confl.toString()))
						}
						if confl.relevant {
							fmt.Println(Magenta("Relevant Conflict detected."))
							fmt.Println(Magenta(confl.toString()))
							fmt.Println(Magenta("Involved Origin ASes: "))
							fmt.Println(Magenta(originCounters[confl.referenceAnnouncement.origin].isp))
							for _, j := range confl.conflictingMessages {
								fmt.Println(Magenta(originCounters[j.origin].isp))
							}
						}
					}
				}

				if countInserted == 100000*countInserted100000 {
					countInserted100000++
					fmt.Println(Green("inserted messages so far: ", countInserted))
				}
				countInserted++

			} else {
				fmt.Println(Red("subnet length = 0. Will not insert: ", m.toString()))
			}
		} else {
			if m.subnet.IP.To4() != nil {
				fmt.Println(Red("IPv4 with more than 32 bits: \n", m.toString()))
			}
			if m.subnet.IP.To16() != nil {
				//[TODO for possible further development]: IPv6 Trie
			}
		}
	} else {
		cleanup()
	}

}
