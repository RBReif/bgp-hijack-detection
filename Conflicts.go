package main

type conflicts struct {
	referenceIPasField    []uint8 //representing the node from which we want to start to find Conflicts
	referenceAnnouncement message //BGP update which triggered a conflict
	conflictingMessages   []message
}

func (conf conflicts) toString() string {
	result := "announcement: " + conf.referenceAnnouncement.toString() + "\n"
	result = result + " the following conflicts were detected: \n"

	for i := 0; i < len(conf.conflictingMessages); i++ {

		result = result + "    " + conf.conflictingMessages[i].toString() + "\n"

	}
	result = result + "\n" + "\n"
	return result

}
