package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type Line struct {
	Epoch   float64
	Db      int
	Command string
	Key     string
	Args    string
}

type Command struct {
	Name string
	Cnt  int
	Sum  float64
	Avg  float64
}

type Commands []Command

type Profile struct {
	Call string
	Args string
	Cnt  int
	Max  float64
	Min  float64
	Avg  float64
	Sum  float64
}

type Profiles []Profile

type ByCnt struct{ Profiles }
type ByMax struct{ Profiles }
type ByAvg struct{ Profiles }

type ByCommand struct{ Commands }
type ByHeavyCommand struct{ Commands }

func (p Profiles) Len() int      { return len(p) }
func (p Profiles) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (c Commands) Len() int      { return len(c) }
func (c Commands) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (p ByCnt) Less(i, j int) bool { return p.Profiles[i].Cnt < p.Profiles[j].Cnt }
func (p ByMax) Less(i, j int) bool { return p.Profiles[i].Max < p.Profiles[j].Max }
func (p ByAvg) Less(i, j int) bool { return p.Profiles[i].Avg < p.Profiles[j].Avg }

func (c ByCommand) Less(i, j int) bool      { return c.Commands[i].Cnt < c.Commands[j].Cnt }
func (c ByHeavyCommand) Less(i, j int) bool { return c.Commands[i].Sum < c.Commands[j].Sum }

const (
	SORT_TYPE_BY_MAX = "max"
	SORT_TYPE_BY_AVG = "avg"
	SORT_TYPE_BY_CNT = "cnt"
)

var (
	filePath    = flag.String("f", "", "redis-cli monitor output file")
	listNum     = flag.Int("n", 10, "Show Slowest Calls Count")
	sortType    = flag.String("s", "max", "Set SlowestCalls Type: max, avg, cnt")
	minCountNum = flag.Int("min", 0, "Show Slowest Calls Count over the minCountNum")

	// regexp
	// refs: https://play.golang.org/p/yl6B1oWtvE
	// 0: line
	// 1: epoch
	// 2; db
	// 3: command
	// 4: command args
	// 5: key
	// 6: args
	lineRegexpRule = `([\d\.]+)\s\[(\d+)\s\d+\.\d+\.\d+\.\d+:\d+]\s"(\w+)"(\s"([^(?<!\\)"]+)"|\s"([^(?<!\\)"]+)"\s(.+)|)`
	lineRegexp     = regexp.MustCompile(lineRegexpRule)

	readFile    *os.File
	monitorLogs Profiles
	commands    Commands

	logIndex  = make(map[string]int)
	logLength = 0
	logCursor = 0

	commandIndex  = make(map[string]int)
	commandLength = 0
	commandCursor = 0

	lineCount = 0
)

func SetProfileIndex(call string) {
	if _, ok := logIndex[call]; ok {
		logCursor = logIndex[call]
	} else {
		logIndex[call] = logLength
		logCursor = logLength
		logLength++
		monitorLogs = append(monitorLogs, Profile{
			Call: call,
			Cnt:  0,
			Min:  0,
			Avg:  0,
			Sum:  0,
		})
	}
}

func SetCommandIndex(command string) {
	if _, ok := commandIndex[command]; ok {
		commandCursor = commandIndex[command]
	} else {
		commandIndex[command] = commandLength
		commandCursor = commandLength
		commandLength++
		commands = append(commands, Command{
			Name: command,
			Cnt:  0,
			Avg:  0,
			Sum:  0,
		})
	}
}

func main() {
	flag.Parse()

	if !strings.Contains(*sortType, SORT_TYPE_BY_MAX) && !strings.Contains(*sortType, SORT_TYPE_BY_AVG) && !strings.Contains(*sortType, SORT_TYPE_BY_CNT) {
		log.Fatal("Please set sort type max, avg or cnt")
	}

	if *filePath != "" {
		rf, err := os.Open(*filePath)
		if err != nil {
			log.Fatal(err)
		}
		readFile = rf
		defer rf.Close()
	} else {
		readFile = os.Stdin
	}

	scanner := bufio.NewScanner(readFile)
	var beforeLine Line

	for scanner.Scan() {
		group := lineRegexp.FindStringSubmatch(scanner.Text())
		if 1 > len(group) {
			continue
		}

		epoch, err := strconv.ParseFloat(group[1], 64)
		if err != nil {
			log.Println(err)
			continue
		}
		db, err := strconv.Atoi(group[2])
		if err != nil {
			log.Println(err)
			continue
		}

		line := Line{
			Epoch:   epoch,
			Db:      db,
			Command: group[3],
			Key:     group[5],
			Args:    group[6],
		}

		var commandTime float64
		if beforeLine.Command != "" {
			commandTime = line.Epoch - beforeLine.Epoch
		} else {
			commandTime = 0
		}

		SetProfileIndex(fmt.Sprintf("%s %s", line.Command, line.Key))
		SetCommandIndex(line.Command)

		if monitorLogs[logCursor].Max < commandTime {
			monitorLogs[logCursor].Max = commandTime
		}

		if commandTime < monitorLogs[logCursor].Min {
			monitorLogs[logCursor].Min = commandTime
		}
		monitorLogs[logCursor].Cnt++
		monitorLogs[logCursor].Sum = monitorLogs[logCursor].Sum + commandTime
		monitorLogs[logCursor].Avg = monitorLogs[logCursor].Sum / float64(monitorLogs[logCursor].Cnt)

		commands[commandCursor].Cnt++
		commands[commandCursor].Sum = commands[commandCursor].Sum + commandTime

		beforeLine = line
		lineCount++
	}

	PrintResult()
}

func PrintTitle(key string) {
	fmt.Println(key)
	fmt.Println("==================================")
}

func PrintResult() {
	PrintTitle("Overall Stats")
	fmt.Printf("LineCount \t %d \n\n", lineCount)
	fmt.Printf("\n")

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', tabwriter.AlignRight)
	PrintTitle("Commands Rate")
	sort.Sort(sort.Reverse((ByCommand{commands})))
	for _, v := range commands {
		fmt.Fprintf(w, "%s\t %d \n", v.Name, v.Cnt)
	}
	w.Flush()
	fmt.Printf("\n")

	PrintTitle("Heavy Commands")
	fmt.Fprintln(w, "Command \tSum(msec)")
	sort.Sort(sort.Reverse((ByHeavyCommand{commands})))
	for _, v := range commands {
		fmt.Fprintf(w, "%s \t %f \n", v.Name, v.Sum)
	}
	w.Flush()
	fmt.Printf("\n")

	PrintTitle("Slowest Calls")
	if strings.Contains(*sortType, SORT_TYPE_BY_AVG) {
		sort.Sort(sort.Reverse((ByAvg{monitorLogs})))
	} else if strings.Contains(*sortType, SORT_TYPE_BY_CNT) {
		sort.Sort(sort.Reverse((ByCnt{monitorLogs})))
	} else {
		sort.Sort(sort.Reverse((ByMax{monitorLogs})))
	}

	fmt.Fprintln(w, "KEY \tCount \tMax(msec) \t Avg(msec)")
	count := 0
	for _, v := range monitorLogs {
		if v.Cnt < *minCountNum {
			continue
		}
		if *listNum < count {
			break
		}
		fmt.Fprintf(w, "%s\t %d \t %f\t %f\n", v.Call, v.Cnt, v.Max, v.Avg)
		count++
	}
	w.Flush()
}
