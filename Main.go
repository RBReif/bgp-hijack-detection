package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

//input
var inputDirectory string //  our initial Routing Information Base and update files
var rib string
var findConflictsInRib bool
var prefixesFileName string
var prefixesFile *os.File

//output
var memProfileFile string
var cpuProfileFile string
var conflictsFileName string
var conflictsFile *os.File
var originsFileName string
var originsFile *os.File
var verbose bool

//live
var liveMode bool
var endLiveStream string
var liveStream string
var risClient string
var buffer int
var writeInterval int

//var routeCollector string

//internal
var flagsString string
var ipv4T ipv4trieRoot
var countInserted int
var countInserted100000 int
var startT time.Time
var stopT time.Time
var mutex *sync.Mutex

var countConflictTriggers int
var countConflictTriggers1000 int
var countConflicts int

func parseFlags() {

	//input
	flag.StringVar(&inputDirectory, "input", "", "If specified, directory containing initial routing information files. Expected filenames: [rib|updates].YYYYMMDD.HHMM{.bz2|.gz}")
	flag.StringVar(&rib, "rib", "", "If specified, we read the specified RIB and all following update files. If not specified the newest RIB in the input directory is used. Expected format: rib.YYYYMMDD.HHMM{.bz2|.gz}")
	flag.BoolVar(&findConflictsInRib, "ribconflicts", false, "If set to true a specified RIB will directly be analysed for conflicts. If set to false (default) only updates (from updates files or from a live feed can trigger conflicts")
	flag.StringVar(&prefixesFileName, "prefixesfile", "input/prefixes", "If specified, a file in which a number of prefixes is listed, for which a conflict needs to be printed out immediately.")

	//output
	flag.StringVar(&cpuProfileFile, "cpuprofile", "output/cp", "Specifies the file to which a CPU profile shall be written to")
	flag.StringVar(&memProfileFile, "memprofile", "output/mp", "Specifies the file to which a memory profile shall be written to")
	flag.StringVar(&conflictsFileName, "conflictsfile", "output/conflicts", "Specifies the file to which found results shall be written to (in Json)")
	flag.StringVar(&originsFileName, "originsfile", "output/origins", "Specifies the file to which frequencies of origin ASes shall be written to (in CSV)")
	flag.BoolVar(&verbose, "verbose", false, "If true we print out found conflicts directly. Defaults to false")
	flag.IntVar(&writeInterval, "interval", 20, "Specifies the interval after which a new originsfile gets written")

	//live
	flag.BoolVar(&liveMode, "live", true, "Indicates if we work in live mode. If in Live mode, input stream has to be specified. If not in live mode, update file has to be specified. Defaults to true")
	flag.StringVar(&liveStream, "stream", "https://ris-live.ripe.net/v1/stream/?format=json", "RIS Live firehose url")
	flag.IntVar(&buffer, "buffer", 10000, "Max depth of Ris messages to queue.")
	flag.StringVar(&risClient, "risclient", "Analysis tool for BGP Hijacks for Summer of Code project of BND", "RIS Live client description")
	flag.StringVar(&endLiveStream, "endlive", "", "If specified, we end the livestream at this time. Expected format: YYYYMMDD.HHMM")
	//flag.StringVar(&routeCollector, "routecollector", "", "If specified only use live stream data from collector with this ID. (expected format: rrcXX). If none is specified, all collectors are included.")

	flag.Parse()

	if endLiveStream != "" {
		s := strings.Split(endLiveStream, ".")
		if len(s) != 2 {
			fmt.Println(Red("ending time in wrong format. Expected format: YYYYMMDD.HHMM"))
			return
		}
		year, _ := strconv.Atoi(s[0][:4])
		month, _ := strconv.Atoi(s[0][4:6])
		day, _ := strconv.Atoi(s[0][6:8])
		hour, _ := strconv.Atoi(s[1][:2])
		minute, _ := strconv.Atoi(s[1][2:4])
		stopT = time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local) //local time zone is used
		fmt.Println(Teal("               current time: ", time.Now().String()))
		fmt.Println(Teal("converted provided end time: ", stopT.String()))
	} else {
		stopT = time.Time{}
	}

	flagsString = "----------------------------------------------------------------------------------------------------------------------------------\n" +
		"Flags parsed:\n input = " + inputDirectory +
		",\n rib = " + rib +
		",\n findConflicts = " + strconv.FormatBool(findConflictsInRib) +
		",\n prefixesFile = " + prefixesFileName +
		"\n" +
		"\n cpuprofile = " + cpuProfileFile +
		",\n memprofile = " + memProfileFile +
		",\n verbose = " + strconv.FormatBool(verbose) +
		",\n conflictsFile = " + conflictsFileName +
		",\n originsFile = " + originsFileName +
		",\n interval = " + strconv.Itoa(writeInterval) +
		",\n" +
		"\n live = " + strconv.FormatBool(liveMode) +
		",\n endlive = " + endLiveStream +
		",\n stream = " + liveStream +
		",\n risclient = " + risClient +
		//", routecollector= " + routeCollector+
		",\n buffer = " + strconv.Itoa(buffer) + "\n" +
		"----------------------------------------------------------------------------------------------------------------------------------\n"

	fmt.Println(Teal(flagsString))
}

func cleanup() {
	fmt.Println(Teal("\n\n----------------------------------------------------------------------------------------------------------------------------------"))
	fmt.Println(Teal("Stopping of program was initiated\n"))
	//PrintMemUsage()

	if memProfileFile != "" {
		f, err := os.Create(memProfileFile)
		if err != nil {
			fmt.Println(Red("could not create Memory Profile file %v", memProfileFile))
			panic(err)
		}
		errorMP := pprof.WriteHeapProfile(f)
		if errorMP != nil {
			fmt.Println(Red("could not create memory profile itself"))
			panic(errorMP)
		}
		errorMPf := f.Close()
		if errorMPf != nil {
			fmt.Println(Red("could not close file in which memory profile is written to"))
		}

	}
	if cpuProfileFile != "" {
		pprof.StopCPUProfile()
	}
	fmt.Println(Teal("Program ran from ", startT, " till ", time.Now()))
	fmt.Println(Green("Inserted messages in total: ", countInserted))
	fmt.Println(Yellow("Peers added in total: ", highestPeerId))
	fmt.Println(White("Messages triggering conflicts: " + strconv.Itoa(countConflictTriggers)))
	fmt.Println(White("Conflicts found: " + strconv.Itoa(countConflicts)))

	fmt.Println()
	printShortSummary()
	fmt.Println()
	fmt.Println(Teal("Writing summary file..."))
	writeOriginFrequencies()
	os.Exit(1)

}

func initialize() {
	fmt.Println(Teal("Initializing IDP BGP Hijack Detection"))

	ourPeers = make([]peer, 0)
	peermapByID = make(map[uint16]*peer)
	peermapByIP = make(map[string]*peer)
	originCounters = make(map[uint32]*originCounter)
	ipv4T = ipv4trieRoot{
		childZero: &ipv4trie{value: 0, representedNet: []uint8{0}},
		childOne:  &ipv4trie{value: 1, representedNet: []uint8{1}},
	}
	readSpecialPrefixes()
	fmt.Println(Teal("Initialization finished at ", time.Now()))

}

func processBGPFiles() {
	files, err := ioutil.ReadDir(inputDirectory) //returnes all files sorted by filename
	if err != nil {
		fmt.Println(Red("could not read specified input directory: ", err))
		return
	}
	fmt.Println(Teal("\n\n----------------------------------------------------------------------------------------------------------------------------------"))

	//no RIB was specified, hence we use the newest one (if a RIB file is there)
	if rib == "" {
		fmt.Println(Teal("No RIB specified. Searching for newest RIB"))
		for i := len(files) - 1; i >= 0; i-- {
			if strings.Contains(files[i].Name(), "rib") {
				fmt.Println(Green("Found newest RIB: ", files[i].Name()))
				rib = files[i].Name()
				break
			}
		}
	}

	//if there is a relevant RIB (the one specified or the newest one) we read the RIB and all following update files
	//if there is no RIB we read all update files
	dateAndTimeStartReading := "0000.0000"
	if rib != "" {
		fmt.Println(Teal("Reading RIB: ", rib))
		e := bgpd.parseRIBAndInsert(inputDirectory + "/" + rib)
		if e != nil {
			fmt.Println(Red("Error while parsing RIB: ", e))
			return
		}
		fmt.Println(Teal("Finished parsing the RIB\n\n"))
		s := strings.Split(rib, ".")
		if !(len(s) > 1) {
			fmt.Println(Red("RIB file in wrong format. Expected format: rib.YYYYMMDD.HHMM{.bz2|.gz}"))
			return
		}
		dateAndTimeStartReading = s[1] + "." + s[2]
		fmt.Println(Teal("Searching for update files representing time intervals after the RIB..."))
	} else {
		fmt.Println(Teal("No RIB found at all. We will now read and parse all updates files"))
	}
	fmt.Println(Teal("\n\n----------------------------------------------------------------------------------------------------------------------------------"))

	for i := 0; i < len(files); i++ {
		if strings.Contains(files[i].Name(), "updates") {
			s := strings.Split(files[i].Name(), ".")
			if !(len(s) > 1) {
				fmt.Println(Red("Updates file in wrong format. Expected format: updates.YYYYMMDD.HHMM{.bz2|.gz}. Got: ", files[i].Name()))
				continue
			}
			dateAndTimeOfFile := s[1] + "." + s[2]
			fmt.Println(Teal("\nFound update File " + dateAndTimeOfFile))
			fmt.Println(Teal("      -->will be inserted: ", dateAndTimeOfFile >= dateAndTimeStartReading))
			if dateAndTimeOfFile >= dateAndTimeStartReading {
				e := messages.parseUpdatesAndInsert(inputDirectory+"/"+files[i].Name(), true)
				if e != nil {
					fmt.Println(Red(e))
				}
				fmt.Println(Teal("Finished parsing and processing of update file ", files[i].Name()))
				fmt.Println()
			}
		}
	}
}

func main() {
	fmt.Println(Teal("Program was started"))
	fmt.Println(Teal("Parsing of flags..."))
	parseFlags()
	startT = time.Now()
	//cpuprofile
	if cpuProfileFile != "" {
		f, err := os.Create(cpuProfileFile)
		if err != nil {
			fmt.Println(Red("could not create CPU Profile file %v", cpuProfileFile))
			panic(err)
		} else {
			fmt.Println(Teal("Started creation of CPU-Profile"))
		}
		er := pprof.StartCPUProfile(f)
		if er != nil {
			fmt.Println(Red("could not create CPU Profile itself"))
		}
		defer pprof.StopCPUProfile()
	}
	fmt.Println(Teal("\nStarting Initialization..."))
	initialize()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
	}()

	var err error
	conflictsFile, err = os.Create(conflictsFileName + ".json")
	if err != nil {
		fmt.Println(Red("could not create JSON file for found conflicts"))
		return
	}
	defer conflictsFile.Close()

	if inputDirectory != "" {
		fmt.Println(Teal("\nStarted parsing of Routeviews..."))
		processBGPFiles()
	}

	if liveMode {
		fmt.Println(Teal("----------------------------------------------------------------------------------------------------------------------------------\n"))

		fmt.Println(Teal("Started Connection to RIPE RIS...\n"))

		mutex = &sync.Mutex{}
		go checkForTimeIntervall(writeInterval)
		for {
			runLivestream()
		}
	}

	cleanup()
}
