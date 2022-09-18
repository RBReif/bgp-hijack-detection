package main

import (
	"encoding/json"
	"fmt"
)

type messageJSON struct {
	Subnet    string `json:"subnet"`
	OriginAS  int    `json:"origin"`
	Timestamp int    `json:"timestamp"`
	Aspath    []int  `json:"aspath"`
}

type ConflictJSON struct {
	ReferenceAnnouncement messageJSON   `json:"referenceAnnouncement"`
	Conflicts             []messageJSON `json:"conflicts"`
}

func convertMessageForJSON(m message) messageJSON {
	mesJSON := messageJSON{
		Subnet:    m.subnet.String(),
		OriginAS:  int(m.origin),
		Timestamp: int(m.timestamp),
		Aspath:    aspathtoIntSlice(m.aspath),
	}
	return mesJSON
}

func convertMessagesForJSON(m []message) []messageJSON {
	messages := make([]messageJSON, len(m))
	for i := 0; i < len(m); i++ {
		messages[i] = convertMessageForJSON(m[i])
	}
	return messages
}

func prepareJSON(c conflicts) {
	conflicts := convertMessagesForJSON(c.conflictingMessages)

	conJSON := ConflictJSON{
		ReferenceAnnouncement: convertMessageForJSON(c.referenceAnnouncement),
		Conflicts:             conflicts,
	}
	writeJson(conJSON)
}

func writeJson(c ConflictJSON) {

	data, err := json.Marshal(c)
	if err != nil {
		fmt.Println(Red("Could not marshal the following conflict(s) as JSON: ", c))
	}
	_, err = conflictsFile.Write(data)
	if err != nil {
		fmt.Println(Red("Could not write to the JSON file for found conflicts: ", string(data)))
		fmt.Println(Red(err))
		fmt.Println()
	} else {
	}

}
