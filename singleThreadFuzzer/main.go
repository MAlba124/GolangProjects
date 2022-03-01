package main

import (
	"fmt"
	"os"
	"bufio"
	"net/http"
	"strings"
)

// Colors
const RED           string  = "\033[31m"
const GREEN         string  = "\033[32m"
const RESET         string  = "\033[0m"
const VERBOSE_COLOR string  = "\033[36m"

func makeRequest(url string, dir string) int {
	resp, err := http.Get(url+dir)

	if err != nil {
		return 0
	}

	defer resp.Body.Close()
	return resp.StatusCode
}

func validateUrl(url string) string {
	var newUrl string
	
	if !strings.HasPrefix(url, "http://") && url[len(url)-1] != '/' {
		newUrl = "http://" + url + "/"
	} else if url[len(url)-1] != '/' {
		newUrl = url + "/"
	} else {
		newUrl = url
	}

	return newUrl
}	

func main() {
	var target string
	var wordList string

	var verbose bool = false

	if len(os.Args) >= 5 && os.Args[1] == "-u" && os.Args[3] == "-wl" {
	    target = validateUrl(os.Args[2])
		wordList = os.Args[4]
	} else {
		fmt.Printf("%sError:%s Missing/wrong arguments!\n", RED, RESET)
		os.Exit(1)
	}

	if len(os.Args) > 5 {
		if os.Args[5] == "-v" {
			verbose = true
			fmt.Printf("[Verbose] %sVerbose mode selected%s\n", VERBOSE_COLOR, RESET)
		}
	}

	if verbose {fmt.Printf("[Verbose] %sOpening word-list:%s %s\n", VERBOSE_COLOR, RESET, wordList)}
    file, error := os.Open(wordList)

    if error != nil {
    	fmt.Printf("%sError:%s %s\n", RED, RESET, error)
    	os.Exit(1)
    } else {
    	fmt.Printf("[Verbose] %sSuccessfully opened word-list:%s %s\n", 
	    		VERBOSE_COLOR, RESET, wordList)
    }
	
	scanner := bufio.NewScanner(file)

    if verbose {fmt.Printf("[Verbose] %sStarting scan%s\n\n", VERBOSE_COLOR, RESET)}
    for scanner.Scan() {
		line := scanner.Text()
		statusCode := makeRequest(target, line)
		if statusCode == 200 {
			fmt.Printf("[ %s%d%s ] Directory found: %s%s\n", GREEN, statusCode, RESET, target, line)
		} else if statusCode == 0 {
			fmt.Printf("%sError:%s [ %d ] StatusCode is 0, did you specify port?\nExiting...\n", RED, RESET, statusCode)
			os.Exit(1)
		}
    }
    fmt.Printf("Scan finnished\n")
}
