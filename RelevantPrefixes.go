package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func readSpecialPrefixes() {
	if prefixesFileName == "" {
		fmt.Println(Teal("No file for special prefixes provided."))
		return
	}

	exists, _ := Exists(prefixesFileName)
	if !exists {
		fmt.Println(Red("File ", prefixesFileName, " does not exist. Continuing without regard for special prefixes."))
		return
	}
	var err error
	prefixesFile, err = os.Open(prefixesFileName)
	if err != nil {
		fmt.Println(Red("Could not open file ", prefixesFileName, ": ", err, ". Continuing without regard for special prefixes"))
	}
	defer prefixesFile.Close()

	scanner := bufio.NewScanner(prefixesFile)
	for scanner.Scan() {
		line := scanner.Text()
		lineArray := strings.Split(line, "/")
		ip := net.ParseIP(lineArray[0])
		if ip == nil {
			fmt.Println(Red("Could not parse IP: ", lineArray[0]))
			continue
		}
		if ip.To4() == nil {
			fmt.Println(Red("IP is not of type IPv4: ", lineArray[0]))
			continue
		}
		atoi, err := strconv.Atoi(lineArray[1])
		if err != nil {
			fmt.Println(Red("could not convert subnetlength to int: ", lineArray[0]))
			continue
		}
		if atoi <= 0 {
			fmt.Println(Red("Subnet length must be at least 1 ", lineArray[0]))
			continue
		}
		_, ipnet, err := net.ParseCIDR(line)
		if err != nil {
			fmt.Println(Red("Could not parse to IPv4 subnet", line))
			continue
		}
		var m message
		m.subnet = *ipnet
		if m.subnet.IP.To4() != nil {

			m.subnetAsBits = convertIPtoBits(m.subnet)
		}
		m.isSpecialPrefix = true
		ipv4T.insert(m)

		fmt.Println(Green("The following prefix was successfully marked as relevant: ", line))

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
