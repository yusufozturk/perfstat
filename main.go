package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"code.google.com/p/go.crypto/ssh/terminal"

	"github.com/pavel-paulau/perfstat/plugins"
)

var header []string
var values []float64

func printHeader() {
	hr := 0
	for _, column := range header {
		fmt.Printf("%s ", column)
		hr += len(column) + 1
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", hr-1))
}

func printValues() {
	for i, column := range header {
		fmtStr := fmt.Sprintf("%%%dv ", len(column))
		fmt.Printf(fmtStr, values[i])
	}
	fmt.Println()
	values = []float64{}
}

func main() {
	cpu := flag.Bool("cpu", false, "enable CPU stats")
	mem := flag.Bool("mem", false, "enable memory stats")

	interval := flag.Int("interval", 1, "sampling interval in seconds")

	quiet := flag.Bool("quiet", false, "disable reporting to stdout")
	perfkeeper := flag.String("perfkeerper", "127.0.0.1:8080", "optional perfkeeper host:port")
	snapshot := flag.String("snapshot", "", "name of perfkeeper snapshot")
	source := flag.String("source", "", "name of perfkeeper snapshot")

	flag.Parse()

	activePlugins := []plugins.Plugin{}
	if *cpu == true {
		activePlugins = append(activePlugins, plugins.NewCPU())
	}
	if *mem == true {
		activePlugins = append(activePlugins, plugins.NewMem())
	}
	if len(activePlugins) == 0 {
		log.Fatalln("Please specify at least one plugin")
	}

	var keeper *Keeper
	if *snapshot != "" && *source != "" {
		keeper = NewKeeper(*perfkeeper, *snapshot, *source)
	} else {
		keeper = nil
	}

	for _, plugin := range activePlugins {
		header = append(header, plugin.GetColumns()...)
	}
	if !*quiet {
		printHeader()
	}

	_, y, err := terminal.GetSize(0)
	if err != nil {
		log.Fatalln(err)
	}

	iterations := 1
	for {
		for _, plugin := range activePlugins {
			values = append(values, plugin.Extract()...)
		}
		if keeper != nil {
			go keeper.Store(header, values)
		}
		if !*quiet {
			printValues()
		}

		iterations ++
		if iterations == y-1 {
			if !*quiet {
				printHeader()
			}
			iterations = 1
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}
