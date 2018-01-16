//Parses perl scripts
package main

import (
	"log"

	"strings"
	"time"
	"unicode"
)

//Path defines file related details
type path struct {
	Name    string
	Scope   string
	Replace string
}

type RootYaml struct {
	Scopes []*ScopeYaml
}

type ScopeYaml struct {
	Name      string
	Functions []*FuncYaml
}

type FuncYaml struct {
	Name     string
	Summary  string
	Example  string
	Argument string
}

type Functions []*API

func (s Functions) Len() int      { return len(s) }
func (s Functions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type FunctionsByName struct{ Functions }

func (s FunctionsByName) Less(i, j int) bool { return s.Functions[i].Function < s.Functions[j].Function }

type Events []*Event

func (s Events) Len() int      { return len(s) }
func (s Events) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type EventsByName struct{ Events }

func (s EventsByName) Less(i, j int) bool { return s.Events[i].Name < s.Events[j].Name }

//API represents an endpoint
type API struct {
	//This is the prefix to a function
	Object string
	//This is the raw functionname, e.g. attacknpc in quest::attacknpc()
	Function string
	//Summary of function
	Summary string
	//Description is pulled from a mapfile
	Description string
	//Scope is object type, e.g. quest in quest::attacknpc()
	Scope string
	//Return is the return type, e.g. bool, void, etc
	Return string
	//Arguments is a list of arguments
	Arguments []*Argument
}

//Argument holds details about arguments
type Argument struct {
	//Name of argument, e.g. item_id
	Name string
	//Type of argument, e.g. int or string
	Type string
	//API holds details about the function the argument is used, mainly for reporting
	API *API
	//Is optional?
	Optional bool
}

type Event struct {
	//Name of event, e.g. EVENT_SAY
	Name string
	//Arguments is a list of arguments
	Arguments []*Argument
}

func main() {

	var err error
	start := time.Now()
	if err = perlGenerate(); err != nil {
		log.Fatalf("Error while generating perl: %s", err.Error())
	}

	log.Println("Finished in", time.Since(start))
}

func getNoun(function string) string {
	function = strings.ToLower(function)

	//first, strip any adjectives
	for k, _ := range adjectives {
		if strings.Index(function, strings.ToLower(k)) == 0 {
			function = function[len(k):]
		}
	}

	//now figure out noun
	for k, v := range nouns {
		if strings.Contains(function, strings.ToLower(k)) {
			return v
		}
	}
	return ""
}

func getAdjective(function string) string {
	function = strings.ToLower(function)
	for k, v := range adjectives {
		if strings.Index(function, strings.ToLower(k)) == 0 {
			return v
		}
	}
	return ""
}

func splitFunctionParts(function string) (parts []string) {
	//first, strip any adjectives
	for k, _ := range adjectives {
		if strings.Index(strings.ToLower(function), strings.ToLower(k)) == 0 {
			function = function[len(k):]
		}
	}

	//try snake
	if strings.Contains(function, "_") { //snake notation
		parts = strings.Split(function, "_")
		return
	}

	//Split it by uppercase
	l := 0
	for s := function; s != ""; s = s[l:] {
		l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
		if l <= 0 {
			l = len(s)
		}

		//The 3 conditionals below is my trying to fix the spaced capitalized words
		if s[:l] == "I" && len(s) > l+1 && s[:l+1] == "ID" {
			//log.Println("Found ID")
			l += 2
		}
		if s[:l] == "M" && len(s) > l+2 && s[:l+2] == "MP3" {
			l += 2
		}
		if s[:l] == "N" && len(s) > l+2 && s[:l+2] == "NPC" {
			//log.Println(s[:l+2])
			l += 3
		}
		parts = append(parts, s[:l])
	}
	return
}
