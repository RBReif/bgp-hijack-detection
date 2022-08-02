package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

//output
var memProfileFile string
var cpuProfileFile string

//live
var liveStream string
var risClient string
var buffer int

//internal
var flagsString string

func parseFlags() {

	//output
	flag.StringVar(&cpuProfileFile, "cpuprofile", "../output/cp", "Specifies the file to which a CPU profile shall be written to")
	flag.StringVar(&memProfileFile, "memprofile", "../output/mp", "Specifies the file to which a memory profile shall be written to")

	//live
	flag.StringVar(&liveStream, "stream", "https://ris-live.ripe.net/v1/stream/?format=json", "RIS Live firehose url")
	flag.IntVar(&buffer, "buffer", 10000, "Max depth of Ris messages to queue.")
	flag.StringVar(&risClient, "risclient", "analysis tool for BGP Hijacks for Summer of Code project of BND", "RIS Live client description")

	flag.Parse()

	flagsString = "Flags parsed: cpuprofile = " + cpuProfileFile + " , memprofile = " + memProfileFile
	fmt.Println(White(flagsString))
}

func cleanup() {
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
	os.Exit(1)

}

//func createDirectory() {   //might not be needed

//	newpath := filepath.Join(".", "output")
//	_ = os.MkdirAll(newpath, os.ModePerm)
//todo errorhandling
//}

func main() {
	fmt.Println("start")

	parseFlags()
	//createDirectory()
	//cpuprofile
	if cpuProfileFile != "" {
		f, err := os.Create(cpuProfileFile)
		if err != nil {
			fmt.Println(Red("could not create CPU Profile file %v", cpuProfileFile))
			panic(err)
		}
		er := pprof.StartCPUProfile(f)
		if er != nil {
			fmt.Println(Red("could not create CPU Profile itself"))
		}
		defer pprof.StopCPUProfile()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
	}()

	for {
		runInLiveMode()
	}

	cleanup()
}
