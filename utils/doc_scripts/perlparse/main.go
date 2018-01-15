//Parses perl scripts
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
)

var isDebug = false

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

//Paths are where every perl file is at
var paths = []*path{
	{
		Name:    "../../../zone/embparser_api.cpp",
		Scope:   "General",
		Replace: "quest",
	},
	{
		Name:  "../../../zone/perl_client.cpp",
		Scope: "Client",
	},
	{
		Name:    "../../../zone/perl_doors.cpp",
		Scope:   "Doors",
		Replace: "door",
	},
	{
		Name:  "../../../zone/perl_entity.cpp",
		Scope: "EntityList",
	},
	{
		Name:  "../../../zone/perl_groups.cpp",
		Scope: "Group",
	},
	{
		Name:  "../../../zone/perl_hateentry.cpp",
		Scope: "HateEntry",
	},
	{
		Name:  "../../../zone/perl_mob.cpp",
		Scope: "Mob",
	},
	{
		Name:  "../../../zone/perl_npc.cpp",
		Scope: "NPC",
	},
	{
		Name:  "../../../zone/perl_object.cpp",
		Scope: "Object",
	},
	{
		Name:    "../../../zone/perl_perlpacket.cpp",
		Scope:   "PerlPacket",
		Replace: "packet",
	},
	{
		Name:  "../../../zone/perl_player_corpse.cpp",
		Scope: "Corpse",
	},
	{
		Name:  "../../../zone/perl_QuestItem.cpp",
		Scope: "QuestItem",
	},
	{
		Name:  "../../../zone/perl_raids.cpp",
		Scope: "Raid",
	},
}

//returnTypes are mapped to identify what sort of return the script does
var returnTypes = map[string]string{
	"boolSV(":   "bool",
	"PUSHu(":    "uint",
	"PUSHi(":    "int",
	"sv_setpv(": "string",
	"PUSHn(":    "double",
}

func main() {

	var err error
	start := time.Now()
	//make an outfile to spit out generated markdowns
	if err = os.Mkdir("out", 0744); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("Failed to make out dir: %s", err.Error())
		}
		err = nil
	}

	sampleYaml := &RootYaml{}

	//functions hold the final function list, all functions get appended so we can group them by scope
	functions := []*API{}

	//iterate all perl files
	for _, path := range paths {
		newFunctions, err := processFile(path)
		if err != nil {
			log.Panicf("Failed to read file: %s", err.Error())
		}

		//we append the newFunctions found from processing the file
		//into functions, for later grouping/processing
		for _, api := range newFunctions {
			functions = append(functions, api)
		}
	}

	log.Println("loaded", len(functions), "functions")

	//outbuffer is grouped by scope
	//I had to do this because not every perl file aligns to scope
	outBuffer := map[string]string{}

	//iterate functions for final output
	for _, api := range functions {
		if api.Scope == "" {
			continue
		}
		if api.Summary == "" {
			api.Summary = "Some summary of function here"
		}
		//prepare a new line
		line := "* [["
		line += fmt.Sprintf("%s(", api.Object)
		//build out arguments
		for _, argument := range api.Arguments {
			if strings.TrimSpace(argument.Name) == "THIS" {
				continue
			}
			if len(strings.TrimSpace(argument.Type)) == 0 {
				line += fmt.Sprintf("%s, ", argument.Name)
			} else {
				line += fmt.Sprintf("%s %s, ", argument.Type, argument.Name)
			}
		}
		//if arguments were shown, remove last ,
		if strings.Contains(line, ",") {
			line = line[0 : len(line)-2]
		}
		//enclose function with a comment of return type
		line += fmt.Sprintf(") # %s", api.Return)
		line += fmt.Sprintf("|Perl-%s-%s]]\n", strings.Title(api.Scope), strings.Title(api.Function))
		//add to outbuffer based on scope
		outBuffer[api.Scope] += line
		isScoped := false
		for _, scope := range sampleYaml.Scopes {
			if scope.Name == api.Scope {
				isScoped = true
			}
		}
		if !isScoped {
			sampleYaml.Scopes = append(sampleYaml.Scopes, &ScopeYaml{
				Name: api.Scope,
			})
		}
		for _, scope := range sampleYaml.Scopes {
			if scope.Name == api.Scope {
				isExists := false
				for _, function := range scope.Functions {
					if function.Name == api.Function {
						isExists = true
						break
					}
				}
				if !isExists {
					summary := ""
					adj := getAdjective(api.Function)
					noun := getNoun(api.Function)
					parts := splitFunctionParts(api.Function)
					if len(parts) > 0 && len(adj) > 0 {
						summary = fmt.Sprintf("%s a %s ", adj, strings.ToLower(scope.Name))

						for _, part := range parts {
							summary += fmt.Sprintf("%s ", strings.ToLower(part))
						}
						summary = summary[0:len(summary)-1] + "."

					} else if len(adj) > 0 && len(noun) > 0 {
						summary = fmt.Sprintf("%s a %s's %s.", adj, strings.ToLower(scope.Name), noun)
					} else {
						summary = fmt.Sprintf("%s.", api.Function)
					}

					arguments := ""
					argCount := 0

					examplePrep := ""
					exampleArgs := ""

					for _, argument := range api.Arguments {
						if argument.Name == "THIS" {
							continue
						}

						exampleType := "1"
						if argument.Type == "string" {
							exampleType = `"test"`
						}
						if strings.Contains(argument.Name, " ") { //this is a bug?
							//fmt.Println("Argument", argument.Name, "in", api.Scope, "for", api.Function, "is jacked up?")
							argument.Name = argument.Name[strings.Index(argument.Name, " ")+1:]
						}

						examplePrep += fmt.Sprintf("my $%s = %s;\n", argument.Name, exampleType)
						exampleArgs += fmt.Sprintf("$%s, ", argument.Name)
						argCount++
						arguments += fmt.Sprintf("%s|%s|%s\n", argument.Name, argument.Type, "")
					}

					if argCount > 0 {
						arguments = "**Name**|**Type**|**Description**\n:---|:---|:---\n" + arguments
						exampleArgs = exampleArgs[0 : len(exampleArgs)-2]
					}

					example := fmt.Sprintf("\n```perl\n%s\n%s(%s); # Returns %s\n```", examplePrep, api.Object, exampleArgs, api.Return)
					if api.Return != "void" {
						example = fmt.Sprintf("\n```perl\n%smy $val = %s(%s);\nquest::say($val); # Returns %s\n```", examplePrep, api.Object, exampleArgs, api.Return)
					}

					function := &FuncYaml{
						Name:     api.Function,
						Summary:  summary,
						Example:  example,
						Argument: arguments,
					}
					scope.Functions = append(scope.Functions, function)
				}
			}
		}
	}

	//iterate outbuffer, which is grouped by scope
	for k, v := range outBuffer {
		if k == "" {
			continue
		}
		//log.Println(k)
		//v = fmt.Sprintf("**Function**|**Summary**\n:-----|:-----\n%s", v)
		if err = ioutil.WriteFile("out/Perl-"+strings.Title(k)+".md", []byte(v), 0744); err != nil {
			err = errors.Wrap(err, "Failed to write file")
			log.Println(err)
		}
	}

	//write new example functions
	sData, err := yaml.Marshal(sampleYaml)
	if err != nil {
		log.Fatalf("Failed to marshal sample: %s", err.Error())
	}
	if err = ioutil.WriteFile("sample.yml", []byte(sData), 0744); err != nil {
		log.Fatalf("Failed to write sample: %s", err.Error())
	}

	fmt.Println("Found", len(sampleYaml.Scopes), "scopes")
	for _, scope := range sampleYaml.Scopes {
		fmt.Println("Found", len(scope.Functions), "functions in", scope.Name)
		for _, function := range scope.Functions {

			buf := fmt.Sprintf("%s\n", function.Summary)
			if len(function.Argument) > 0 {
				buf += fmt.Sprintf("### Arguments\n%s\n", function.Argument)
			}
			buf += fmt.Sprintf("### Example\n%s", function.Example)
			err = ioutil.WriteFile(fmt.Sprintf("out/Perl-%s-%s.md", strings.Title(scope.Name), function.Name), []byte(buf), 0744)
			if err != nil {
				log.Fatalf("Failed to write file %s %s: %s", scope.Name, function.Name, err.Error())
			}
		}
	}

	log.Println("Finished in", time.Since(start))
}

func processFile(path *path) (functions []*API, err error) {

	inFile, err := os.Open(path.Name)
	if err != nil {
		err = errors.Wrap(err, "Failed to open file")
	}
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	arguments := map[string][]*Argument{}
	reg, err := regexp.Compile(`\]+|\[+|\?+`)
	if err != nil {
		err = errors.Wrap(err, "Failed to compile regex")
		return
	}
	regType, err := regexp.Compile(`(unsigned long|long|int32|bool|uint[0-9]+|int|auto|float|unsigned int|char[ \*]).+([. a-zA-Z]+=)`)
	if err != nil {
		err = errors.Wrap(err, "Failed to compile type regex")
		return
	}

	lastArguments := []*Argument{}
	lastAPI := &API{}
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		key := ""
		line := scanner.Text()
		isDebug = false
		if len(line) < 1 {
			continue
		}

		//see if the line contains any valid return types
		for key, val := range returnTypes {
			if strings.Contains(line, key) {
				lastAPI.Return = val
				break
			}
		}

		if len(lastArguments) > 0 { //existing args to parse

			for i, argument := range lastArguments {
				key = fmt.Sprintf("ST(%d)", i)
				if strings.Contains(line, key) {
					if strings.Contains(lastAPI.Function, "attacknpc") {
						log.Println("found argument for", i, lastAPI.Function)
					}
					if argument.Type != "" {
						continue
					}

					match := regType.FindStringSubmatch(line)
					if len(match) < 2 {
						continue
					}

					//key = `int`
					//function = line[strings.Index(line, key)+len(key):]
					newType := ""

					v := strings.TrimSpace(match[1])

					switch v {
					case "int":
						newType = "int"
					case "int32":
						newType = "int"
					case "float":
						newType = "float"
					case "unsigned int":
						newType = "uint"
					case "uint32":
						newType = "uint"
					case "uint8":
						newType = "uint"
					case "uint":
						newType = "uint"
					case "bool":
						newType = "bool"
					case "uint16":
						newType = "uint"
					case "long":
						newType = "long"
					case "unsigned long":
						newType = "unsigned long"
					default:
						if strings.Contains(v, "auto") {
							if strings.Contains(line, "glm::vec4") {
								newType = "float"
							}
						}
						if strings.Contains(v, "char") {
							newType = "string"
						}
					}
					if newType == "" {
						log.Printf(`Unknown type: "%s" on line %d`, newType, lineNum)
					}
					//log.Println("Found arg type", newType, "on index", i, argument.Name)
					lastArguments[i].Type = newType
				}
			}
		}

		function := ""

		argLine := ""
		args := []string{}
		//Find line
		key = `Perl_croak(aTHX_ "Usage:`
		if strings.Contains(line, key) {
			function = line[strings.Index(line, key)+len(key):]
		}

		for _, argument := range lastArguments {
			arguments[argument.Name] = append(arguments[argument.Name], argument)
		}

		lastArguments = []*Argument{}

		//Trim off the endings
		key = `");`
		if strings.Contains(function, key) {
			function = function[0:strings.Index(function, key)]
		}
		//Strip out the arguments
		key = `(`
		if strings.Contains(function, key) {
			argLine = function[strings.Index(function, key)+len(key):]
			function = function[0:strings.Index(function, key)]
			key = `)`
			if strings.Contains(argLine, key) {
				argLine = argLine[:strings.Index(argLine, key)]
			}
			key = `=`
			if strings.Contains(argLine, key) {
				argLine = argLine[:strings.Index(argLine, key)]
			}

		}
		key = `,`
		argLine = strings.TrimSpace(argLine)

		if strings.Contains(argLine, key) { //there is a , in the argument list
			args = strings.Split(argLine, key)
		} else { //no , in argument list, look for single one
			if len(argLine) > 0 {
				args = []string{
					argLine,
				}
			}
		}

		if len(function) < 1 {
			continue
		}

		function = strings.TrimSpace(function)

		newArgs := []string{}
		for j, _ := range args {
			args[j] = strings.TrimSpace(args[j])
			if len(args[j]) == 0 {
				continue
			}
			newArgs = append(newArgs, args[j])
		}

		if lastAPI != nil {
			isNew := true
			for _, oldFunc := range functions {
				if oldFunc.Function == lastAPI.Function {
					if len(oldFunc.Arguments) > len(lastAPI.Arguments) {
						//log.Println("Skipping", oldFunc, "since less arguments")
						isNew = false
					}
				}
			}
			if isNew {
				functions = append(functions, lastAPI)
			}
		}
		lastAPI = &API{
			Function: function,
		}

		for _, arg := range newArgs {
			isOptional := false
			if strings.Contains(arg, "]") {
				isOptional = true
			}
			arg = reg.ReplaceAllString(arg, "")
			argType, _ := knownTypes[arg]
			argument := &Argument{
				Name:     arg,
				Type:     argType,
				API:      lastAPI,
				Optional: isOptional,
			}

			lastArguments = append(lastArguments, argument)
		}
		lastAPI.Arguments = lastArguments
	}

	fmt.Printf("==========%s==========\n", path.Scope)
	foundCount := 0
	failCount := 0
	for key, val := range arguments {
		isMissing := false
		line := ""
		line = fmt.Sprintf("%s used by %d functions:", key, len(val))
		for _, fnc := range val {
			line += fmt.Sprintf("%s(%s %s), ", fnc.API.Function, fnc.Type, key)
			if fnc.Type == "" {
				isMissing = true
			}
		}
		if isMissing {
			fmt.Println(line)
			failCount++
		} else {
			foundCount++
		}
	}
	log.Println(foundCount, "functions properly identified,", failCount, "have errors")

	for _, api := range functions {
		if len(api.Function) == 0 {
			continue
		}

		api.Function = strings.TrimSpace(api.Function)

		if api.Function == `%s` {
			continue
		}

		if api.Return == "" {
			api.Return = "void"
		}

		if path.Replace == "" {
			path.Replace = strings.ToLower(path.Scope)
		}

		//Figure out object
		if path.Scope != "General" {
			api.Object = api.Function
			if strings.Contains(api.Object, path.Scope+"::") {
				api.Object = strings.Replace(api.Object, path.Scope+"::", strings.ToLower(path.Scope)+"->", -1)
				api.Object = "$" + strings.TrimSpace(api.Object)
			} else {
				api.Object = "$" + strings.TrimSpace(strings.ToLower(api.Object)) + "->"
			}
			if strings.Contains(api.Object, "::") {
				api.Object = strings.Replace(api.Object, "::", "->", -1)
			}
		} else {
			if !strings.Contains(api.Object, path.Replace) {
				api.Object = "quest::" + api.Object
			}
		}

		//Strip out object from function
		if strings.Contains(api.Function, "::") {
			api.Function = api.Function[strings.Index(api.Function, "::")+2:]
		}
		if strings.Contains(api.Function, "->") {
			api.Function = api.Function[0:strings.Index(api.Function, "->")]
		}
		if strings.Contains(api.Function, "$") {
			api.Function = api.Function[1:]
		}

		//Figure out scope
		if strings.Contains(api.Object, "::") {
			api.Scope = api.Object[0:strings.Index(api.Object, "::")]
		}
		if strings.Contains(api.Object, "->") {
			api.Scope = api.Object[0:strings.Index(api.Object, "->")]
		}

		if strings.Contains(api.Scope, "$") {
			api.Scope = api.Scope[strings.Index(api.Scope, "$")+1:]
		}
	}

	return
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

var adjectives = map[string]string{
	"get":      "gets",
	"send":     "sends",
	"set":      "sets",
	"teleport": "teleports",
	"is":       "is",
	"play":     "plays",
	"add":      "adds",
}

var nouns = map[string]string{
	"taskid":     "[task](Task)",
	"account_id": "[account](Task)",
	"accountid":  "[account](Task)",
}

//These are known parameter types
var knownTypes = map[string]string{
	"activity_id":               "uint",
	"alt_mode":                  "bool",
	"anim_num":                  "int",
	"best_z":                    "float",
	"buttons":                   "int",
	"channel_id":                "int",
	"char_id":                   "int",
	"charges":                   "int",
	"class_id":                  "int",
	"client_name":               "string",
	"color":                     "int",
	"color_id":                  "int",
	"condition_id":              "int",
	"copper":                    "int",
	"count":                     "int",
	"debug_level":               "int",
	"decay_time":                "int",
	"dest_heading":              "float",
	"dest_x":                    "float",
	"dest_y":                    "float",
	"dest_z":                    "float",
	"in_lastname":               "string",
	"distance":                  "int",
	"door_id":                   "int",
	"value":                     "int",
	"cost":                      "int",
	"slot":                      "int",
	"type":                      "int",
	"iSendToSelf":               "int",
	"iFromDB":                   "bool",
	"duration":                  "int",
	"effect_id":                 "int",
	"elite_material_id":         "int",
	"enforce_level_requirement": "bool",
	"explore_id":                "uint",
	"faction_value":             "int",
	"fade_in":                   "int",
	"fade_out":                  "int",
	"fadeout":                   "uint",
	"firstname":                 "string",
	"format":                    "string",
	"from":                      "string",
	"gender_id":                 "int",
	"gold":                      "int",
	"grid_id":                   "int",
	"guild_rank_id":             "int",
	"heading":                   "float",
	"hero_forge_model_id":       "int",
	"ignore_quest_update":       "bool",
	"instance_id":               "int",
	"int_unused":                "int",
	"int_value":                 "int",
	"is_enabled":                "bool",
	"is_strict":                 "bool",
	"item_id":                   "int",
	"key":                       "string",
	"language_id":               "int",
	"lastname":                  "string",
	"leader_name":               "string",
	"level":                     "int",
	"link_name":                 "string",
	"macro_id":                  "int",
	"max_level":                 "int",
	"max_x":                     "float",
	"max_y":                     "float",
	"max_z":                     "float",
	"message":                   "string",
	"milliseconds":              "int",
	"min_level":                 "int",
	"min_x":                     "float",
	"min_y":                     "float",
	"min_z":                     "float",
	"name":                      "string",
	"new_hour":                  "int",
	"new_min":                   "int",
	"node1":                     "int",
	"node2":                     "int",
	"npc_id":                    "int",
	"npc_type_id":               "int",
	"object_type":               "int",
	"options":                   "int",
	"platinum":                  "int",
	"popup_id":                  "int",
	"priority":                  "int",
	"quantity":                  "int",
	"race_id":                   "int",
	"remove_item":               "bool",
	"requested_id":              "int",
	"reset_base":                "bool",
	"saveguard":                 "bool",
	"seconds":                   "int",
	"send_to_world":             "bool",
	"signal_id":                 "int",
	"silent":                    "bool",
	"silver":                    "int",
	"size":                      "int",
	"spell_id":                  "int",
	"stat_id":                   "int",
	"str_value":                 "string",
	"subject":                   "string",
	"target_enum":               "string",
	"target_id":                 "int",
	"task":                      "int",
	"task_id":                   "uint",
	"task_id1":                  "int",
	"task_id10":                 "int",
	"task_id2":                  "int",
	"task_set":                  "int",
	"taskid":                    "int",
	"taskid1":                   "int",
	"taskid2":                   "int",
	"taskid3":                   "int",
	"taskid4":                   "int",
	"teleport":                  "int",
	"temp":                      "int",
	"texture_id":                "int",
	"theme_id":                  "int",
	"update_world":              "int",
	"updated_time_till_repop":   "uint",
	"version":                   "int",
	"wait_ms":                   "int",
	"window_title":              "string",
	"x":                         "float",
	"y":                         "float",
	"z":                         "float",
	"zone_id":                   "int",
	"zone_short":                "string",
	`task_id%i`:                 "int",
}
