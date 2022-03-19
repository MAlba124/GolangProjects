package main

import (
	"fmt"
	"flag"
	"os"
	"errors"
	"bufio"
)

type Data struct {
	passSearch    bool
	hashSearch    bool
	password    string
	hash        string
	rockYouPath string
	verbose       bool
	color         bool
	search        bool
}

var (
	green string =  "\033[32m"
	red string   =  "\033[31m"
	reset string =  "\033[0m"
	cyan string  =  "\033[36m"
)

func main() {

	var d Data
	d.init()

	if d.color != true {
		green  =  ""
		red    =  ""
		reset  =  ""
		cyan   =  ""
	}

	if len(os.Args) < 2 {
		fmt.Printf("%s[ ! ] No arguments provided%s\n", red, reset)
		return
	}

	if d.search {
		if fileExists(d.rockYouPath) != true {
			var searchRY string
			fmt.Printf("%s[ ! ] RockYou was not found!%s %s|%s Do you want to search you home directory for it? %s[ y/n ]%s > ", red, reset, cyan, reset, cyan, reset)
			fmt.Scanf("%s", &searchRY)
			if searchRY == "y" || searchRY == "Y" {
				fmt.Printf("%s[ * ] Searching for rock you%s\n", cyan, reset)
			}
		} else {
			if d.passSearch {
				fmt.Printf("%s[ * ] Searching in RockYou%s\r", green, reset)
				line, retErr := d.searchRY()

				if retErr != 0 && retErr != 1 {
					fmt.Printf("%s[ - ] Failed to search for password%s\n", red, reset)
					return
				} else if retErr == 0 {
					fmt.Printf("%s[ ! ]%s Your password was found in RockYou on line: %s%d |%s Please change it imideatley!\n", red, reset, cyan, line, reset)
					return
				} else if retErr == 1 {
					fmt.Printf("%s[ ! ]%s Your password was NOT found in RockYou!\n", green, reset)
					return
				}
			}
		}
	} else {
		fmt.Printf("%s[ ! ] No password/hash provided%s\n", red, reset)
		return
	}
}

func fileExists(file string) bool {

	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (d *Data)searchRY() (int, int) {


	var line int = 0

	file, err := os.Open(d.rockYouPath)
		if err  != nil {
		return line, -1
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if scanner.Text() == d.password {
			return line, 0
		}
		line++
	}
	return 0, 1
}

func (d *Data)init() {

	var pS, hS, color, verbose bool
	var passWd, hash, ryPath string

    flag.StringVar(&passWd, "p", "", "Specify password to search for")
    flag.StringVar(&hash, "ha", "", "Specify password hash")
 	flag.StringVar(&ryPath, "ryp", "", "Specify RockYou path")
 	
 	flag.BoolVar(&verbose, "v", false, "Specify verbose mode")
 	flag.BoolVar(&color, "c", true, "Disable colored output")
    
    flag.Parse()

    if len(passWd) < 1 {
    	pS = false
    } else {
    	pS = true
    }
    if len(hash) < 1 {
    	hS = false
    } else {
    	hS = true
    }

    d.rockYouPath = ryPath
    d.passSearch = pS
    d.hashSearch = hS
    d.color = color
    d.verbose = verbose

    if pS != true || hS != true {
    	d.search = false
    }
    if pS {
    	d.password = passWd
    	d.search = true	
    }
    if hS {
    	d.hash = hash
    	d.search = true	
    }
}
