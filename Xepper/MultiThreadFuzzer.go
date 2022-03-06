package main

import (
	"os"
	"strings"
	"fmt"
	"sync"
	"io"
	"errors"
	"flag"
	"syscall"
	"net/http"
	"math"
	"time"
	"os/signal"
	"bufio"
)

type Target struct {
	wordlist  string
	url       string
	verbose     bool
	quiet       bool
	lines        int
	threads      int
	nocolor     bool
}

type Res struct {
	baseUrl  string
	dir      string
	thread      int
	size      int64
	statusCode  int
	err        bool
	errMsg   string
}

var (
	green string =  "\033[32m"
	red string   =  "\033[31m"
	reset string =  "\033[0m"
)

func main() {

	if len(os.Args) <= 1 {
		fmt.Printf("Missing arguments! -h for help\n")
		os.Exit(0)
	}

	var target Target
	var exit bool
	
	target, exit = argParser()

	if exit {
		os.Exit(0)
	}
	
	fmt.Printf("Starting Xepper v0.0.1...\n\n")

	if target.nocolor {
		green  =  ""
		red    =  ""
		reset  =  ""
	}

	if !fileExists(target.wordlist) {
		fmt.Printf("Wordlist does not exist\n")
		os.Exit(1)
	}

	if target.verbose {
		fmt.Printf("opening file: %s\n", target.wordlist)
	}
	file, error := os.Open(target.wordlist)
	if error != nil {
		fmt.Printf("Error opening file %s\n", target.wordlist)
		os.Exit(1)
	}

	target.countLines(file)

	if target.lines < 100 && target.threads > 1 {
		fmt.Printf("Provided wordlist is under 100 lines,\nthreads is automaticly set to 1\n")
		target.threads = 1
	}

	file.Seek(0, 0)

	target.validateUrl(target.url)

	target.printBanner()

	if target.threads == 1 {
		fmt.Printf("Under construction\n")
		//target.run(0, target.lines, 0, printChannel)
	} else {
		exit := make(chan bool)

		var wg sync.WaitGroup
		
		interval := math.Round(float64(target.lines) / float64(target.threads))

		if target.verbose {
			fmt.Printf("Creating Goroutines and listener...\n")
		}
 
		ki := make(chan os.Signal)
		signal.Notify(ki, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-ki
			fmt.Printf("\nKeyboard interupt, closing all threads. Stand by!\n")
			for i := 1; i <= target.threads+2; i++ {
				exit<-true
			}
			close(ki)
		}()

		printChannel := make(chan Res, 10000)
		go func(pc chan Res, e chan bool) {
			printer(pc, e)
		}(printChannel, exit)

		countChannel := make(chan int, 1000)
		go func(cc chan int, e chan bool) {
			target.counter(cc, e)
		}(countChannel, exit)

		for i := 0; i <= target.threads; i++ {
			wg.Add(1)
			start := i * int(interval)
			end := (start + int(interval)) - 1
			go func(index int, start int, end int, pc chan Res, e chan bool, cc chan int) {
				target.run(start, end, index, pc, e, cc)
				wg.Done()
			}(i, start, end, printChannel, exit, countChannel)
			if target.verbose {
				fmt.Printf("Goroutine created\n")
			}
		}
		wg.Wait()
		close(printChannel)
		close(exit)
	}
	fmt.Printf("Exiting...\n")
}

func argParser() (Target, bool) {

	var target Target
	var url, wordlist string
	var verbose, quiet, color bool
	var threads int

    flag.StringVar(&url, "u", "", "Specify target url url")
    flag.StringVar(&wordlist, "w", "", "Specify directory wordlist")
 	flag.BoolVar(&verbose, "v", false, "Specify verbose mode")
 	flag.BoolVar(&quiet, "q", false, "Disable printing the banner")
    flag.IntVar(&threads, "t", 1, "Specify the amount of threads")
    flag.BoolVar(&color, "nocolor", false, "Disable colored output")
    
    flag.Parse()

    target.wordlist = wordlist
    target.url      =      url
    target.verbose  =  verbose
    target.threads  =  threads
    target.quiet    =    quiet
    target.nocolor  =    color

    return target, false
}

func fileExists(wordlist string) bool {

	if _, err := os.Stat(wordlist); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func readLine(file io.Reader, line int) (string, error) {

	var cLine int = 0;
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if cLine == line {
			return scanner.Text(), nil
		}
		cLine++
	}
	return " ", nil
}

func (t *Target)printBanner() {

	if !t.quiet {
		// Banner
		fmt.Printf("\n           ,/\n         ,'/\t Target: %s\n       ,' /\n     ,'  /_____, Wordlist: %s\n   .'____    ,'\n        /  ,'\t Words: %d\n       / ,'\n      /,'\t Threads: %d\n     /'\n\n", t.url, t.wordlist, t.lines, t.threads)
	} else {
		fmt.Printf("Target: %s\nWordlist: %s\nWords: %d\nThreads: %d\n\n", 
			t.url, t.wordlist, t.lines, t.threads)
	}
}

func (t *Target)countLines(file io.Reader) {

	var clines int
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		clines++;
	}
	t.lines = clines
}

func (t *Target)run(start int, end int, routine int, pc chan Res, e chan bool, cc chan int) {

	file, err := os.Open(t.wordlist)
		if err  != nil {
		return
	}

	var cLine int = 0
	scanner := bufio.NewScanner(file)
	client := &http.Client{
		Timeout: 6 * time.Second,
	}
	
	count := 0
	for scanner.Scan() {
		select {
		case <-e:
			return
		default:	
			if cLine >= end {
				return
			}
			if cLine >= start && cLine <= end {
				currWord := scanner.Text()
				word := strings.TrimSpace(currWord)
 				if strings.HasPrefix(word, "#") || len(word) == 0 {
					continue 
				}
				statusCode, size := t.makeRequest(word, client)
				if statusCode != 404 && statusCode != 0 {
					pc <- Res{
							baseUrl:         t.url,
							dir:              word,
							thread:        routine,
							size:             size,
							statusCode: statusCode,
							err:             false,
							errMsg:             "",
					}
				} else if statusCode == 0 {
					pc <- Res{
							baseUrl:         t.url,
							dir:              word,
							thread:        routine,
							size:             size,
							statusCode: statusCode,
							err:              true,
							errMsg:      "Exiting",
					}
				}
			}
			count++
			if count > 100 {
				cc<-100
				count = 0
			}
			cLine++
		}
	}
}

func (t *Target)makeRequest(dir string, client *http.Client) (int, int64) {

	request, err := http.NewRequest("GET", t.url+dir, nil)
	if err != nil {
		return 0, 0
	}
	request.Header.Set("User-Agent", "Xepper")

	resp, err := client.Do(request)
	if err != nil {
		return 0, 0
	}

	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, 0
	} else {
		size := int64(len(respBody))
		return resp.StatusCode, size
	}

	return resp.StatusCode, 0
}

func (t *Target)validateUrl(oUrl string) {
	
	if !strings.HasPrefix(oUrl, "http://") && oUrl[len(oUrl)-1] != '/' {
		t.url = fmt.Sprintf("http://%s/", oUrl)
	} else if !strings.HasPrefix(oUrl, "http://") {
		t.url = fmt.Sprintf("http://%s", oUrl)
	}else if oUrl[len(oUrl)-1] != '/' {
		t.url = fmt.Sprintf("%s/", oUrl)
	} else {
		t.url = oUrl
	}
}

func printer(printChannel chan Res, e chan bool) {

	for {
		select{
		case <-e:
			return
		default:
			message := <- printChannel
			if message.err {
				fmt.Printf("Error occurred in thread %d: %s\n", message.thread, message.errMsg)
			} else {
				fmt.Printf("[ Statuscode: %s%d%s ] [ Size: %s%d%s ] %s%s\n", green, message.statusCode, reset, green, message.size, reset, message.baseUrl, message.dir)
			}
		}
	}
}

// This is not working propperly
func (t *Target)counter(c chan int, e chan bool) {

	num := 0
	for {
		select{
		case <-e:
			return
		case <-c:
			num = num + 100
			fmt.Printf("\r%d / %d         ", num, t.lines)
		default:
			continue
		}
	}
}
