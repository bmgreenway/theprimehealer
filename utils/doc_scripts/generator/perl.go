package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
)

func perlGenerate() (err error) {

	//make an outfile to spit out generated markdowns
	if err = os.Mkdir("out", 0744); err != nil {
		if !os.IsExist(err) {
			err = errors.Wrap(err, "Failed to make out dir")
			return
		}
		err = nil
	}

	//functions hold the final function list, all functions get appended so we can group them by scope
	functions := []*API{}
	events := []*Event{}

	//iterate all perl files
	for _, path := range perlPaths {
		newFunctions := []*API{}
		newEvents := []*Event{}
		newFunctions, newEvents, err = perlProcessFile(path)
		if err != nil {
			err = errors.Wrap(err, "Failed to read file")
			return
		}

		//we append the newFunctions found from processing the file
		//into functions, for later grouping/processing
		for _, api := range newFunctions {
			functions = append(functions, api)
		}

		for _, event := range newEvents {
			events = append(events, event)
		}
	}

	log.Println("loaded", len(functions), "functions")

	//functionBuffer is grouped by scope
	//I had to do this because not every perl file aligns to scope

	sort.Sort(FunctionsByName{functions})

	sort.Sort(EventsByName{events})

	functionBuffer, eventBuffer, sampleYaml, err := perlGroupAndPrepareFunctions(functions, events)
	if err != nil {
		err = errors.Wrap(err, "Failed to prepare and group functions")
		return
	}
	if err = perlWriteWikiPages(functionBuffer, eventBuffer, sampleYaml, events); err != nil {
		err = errors.Wrap(err, "Failed to write wiki pages")
		return
	}

	return
}

func perlProcessFile(path *path) (functions []*API, events []*Event, err error) {

	var index int
	inFile, err := os.Open(path.Name)
	if err != nil {
		err = errors.Wrap(err, "Failed to open file")
		return
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
	//since events are in cases
	lastEvents := []*Event{}

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		key := ""
		line := scanner.Text()
		if len(line) < 1 {
			continue
		}

		//See if line has any event info
		index = strings.Index(line, "case EVENT")
		if index > 0 {
			if len(lastEvents) > 0 && len(lastEvents[0].Arguments) > 0 {
				for _, event := range lastEvents {
					events = append(events, event)
				}
				//flush
				lastEvents = []*Event{}
			}

			event := &Event{}
			event.Name = line[index+5:]
			index = strings.Index(event.Name, ":")
			if index > 0 {
				event.Name = event.Name[0:index]
			}
			lastEvents = append(lastEvents, event)
			continue
		}

		index = strings.Index(line, `ExportVar(package_name.c_str(), "`)
		if index > 0 {
			arg := &Argument{}

			arg.Name = line[index+33:]
			index = strings.Index(arg.Name, `"`)
			if index > 0 {
				arg.Name = arg.Name[0:index]
			}
			for _, event := range lastEvents {
				event.Arguments = append(event.Arguments, arg)
			}
			continue
		}

		//see if the line contains any valid return types
		for key, val := range perlReturnTypes {
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
		index = strings.Index(line, key)
		if index > 0 {
			function = line[index+len(key):]
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
			argType, _ := perlKnownTypes[arg]
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

	if len(lastEvents) > 0 {
		for _, event := range lastEvents {
			events = append(events, event)
		}
	}

	fmt.Printf("==========%s==========\n", path.Scope)
	foundCount := 0
	failCount := 0
	for key, val := range arguments {
		if key == "THIS" {
			continue
		}
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
			api.Object = "$" + strings.ToLower(api.Function)

			if strings.Contains(strings.ToLower(api.Object), "hate") {
				fmt.Println(api.Object)
			}
			api.Object = strings.Replace(api.Object, "entitylist", "entity_list", -1)
			api.Object = strings.Replace(api.Object, "hateentry", "hate_entry", -1)
			index = strings.Index(api.Object, "::")
			if index > 0 {
				api.Object = api.Object[0:index] + "->"
			}
			index = strings.Index(api.Object, "->")
			if index > 0 {
				api.Object = api.Object[0 : index+2]
			}
			/*if strings.Contains(api.Object, path.Scope+"::") {
				api.Object = strings.Replace(api.Object, path.Scope+"::", strings.ToLower(path.Scope)+"->", -1)
				api.Object = "$" + strings.TrimSpace(api.Object)
			} else {
				api.Object = "$" + strings.TrimSpace(strings.ToLower(api.Object)) + "->"
			}
			if strings.Contains(api.Object, "::") {
				api.Object = strings.Replace(api.Object, "::", "->", -1)
			}*/
		} else {
			if !strings.Contains(api.Object, path.Replace) {
				api.Object = "quest::"
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
		index = strings.Index(api.Object, "::")
		if index > 0 {
			api.Scope = api.Object[0:index]
		}
		index = strings.Index(api.Object, "->")
		if index > 0 {
			api.Scope = api.Object[0:index]
		}

		if strings.Contains(api.Scope, "$") {
			api.Scope = api.Scope[strings.Index(api.Scope, "$")+1:]
		}
		api.Scope = strings.Title(api.Scope)
		//manually override weirdly spelled ones
		if strings.ToLower(api.Scope) == "entity_list" {
			api.Scope = "EntityList"
		}
		if strings.ToLower(api.Scope) == "hate_entry" {
			api.Scope = "HateEntry"
		}
		if strings.ToLower(api.Scope) == "npc" {
			api.Scope = "NPC"
		}
	}

	return
}

func perlGroupAndPrepareFunctions(functions []*API, events []*Event) (functionBuffer map[string]string, eventBuffer map[string]string, sampleYaml *RootYaml, err error) {
	functionBuffer = make(map[string]string)
	eventBuffer = make(map[string]string)
	sampleYaml = &RootYaml{}

	for _, event := range events {
		line := fmt.Sprintf("* [[%s|Perl-%s]]\n", event.Name, event.Name)
		eventBuffer[""] += line
	}

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
		line += fmt.Sprintf("%s%s(", api.Object, api.Function)
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
		line += fmt.Sprintf("|Perl-%s-%s]]\n", api.Scope, strings.Title(api.Function))
		//add to functionBuffer based on scope
		functionBuffer[api.Scope] += line
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

						if strings.TrimSpace(argument.Name) != "..." {
							examplePrep += fmt.Sprintf("my $%s = %s;\n", argument.Name, exampleType)
							exampleArgs += fmt.Sprintf("$%s, ", argument.Name)
						} else {
							exampleArgs += fmt.Sprintf("%s, ", argument.Name)
						}

						argCount++
						arguments += fmt.Sprintf("%s|%s|%s\n", argument.Name, argument.Type, "")
					}

					if argCount > 0 {
						arguments = "**Name**|**Type**|**Description**\n:---|:---|:---\n" + arguments
						exampleArgs = exampleArgs[0 : len(exampleArgs)-2]
					}

					example := fmt.Sprintf("\n```perl\n%s\n%s%s(%s); # Returns %s\n```", examplePrep, api.Object, api.Function, exampleArgs, api.Return)
					if api.Return != "void" {
						example = fmt.Sprintf("\n```perl\n%smy $val = %s%s(%s);\nquest::say($val); # Returns %s\n```", examplePrep, api.Object, api.Function, exampleArgs, api.Return)
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
	return
}

func perlWriteWikiPages(functionBuffer map[string]string, eventBuffer map[string]string, sampleYaml *RootYaml, events []*Event) (err error) {

	for _, v := range eventBuffer {
		v += fmt.Sprintf("\n\nGenerated On %s", time.Now().Format(time.RFC3339))
		if err = ioutil.WriteFile("out/Perl-Events.md", []byte(v), 0744); err != nil {
			err = errors.Wrap(err, "Failed to write file")
			log.Println(err)
		}
	}

	for _, event := range events {

		argLine := ""
		buf := fmt.Sprintf("%s\n", event.Name)
		if len(event.Arguments) > 0 {
			buf += fmt.Sprintf("### Exports\n**Name**|**Type**|**Description**\n:-----|:-----|:-----\n")
			for _, arg := range event.Arguments {
				if arg.Name == "" {
					continue
				}
				if arg.Type == "" {
					arg.Type = "int"
				}

				buf += fmt.Sprintf("%s|%s|\n", arg.Name, arg.Type)
				argLine += fmt.Sprintf("	quest::say($%s); # returns %s\n", arg.Name, arg.Type)
			}

		}
		buf += fmt.Sprintf("### Example\n")
		buf += fmt.Sprintf("```perl\nsub %s {\n%s}\n```", event.Name, argLine)
		buf += fmt.Sprintf("\n\nGenerated On %s", time.Now().Format(time.RFC3339))
		err = ioutil.WriteFile(fmt.Sprintf("out/Perl-%s.md", strings.Title(event.Name)), []byte(buf), 0744)
		if err != nil {
			err = errors.Wrapf(err, "Failed to write file %s", event.Name)
			return
		}
	}

	//iterate functionBuffer, which is grouped by scope
	for k, v := range functionBuffer {
		if k == "" {
			continue
		}
		v += fmt.Sprintf("\n\nGenerated On %s", time.Now().Format(time.RFC3339))
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
		err = errors.Wrap(err, "Failed to marshal sample")
		return
	}
	if err = ioutil.WriteFile("perlsample.yml", []byte(sData), 0744); err != nil {
		err = errors.Wrap(err, "Failed to write sample")
		return
	}

	fmt.Println("Found", len(sampleYaml.Scopes), "scopes")
	for _, scope := range sampleYaml.Scopes {
		fmt.Println("Found", len(scope.Functions), "functions in", scope.Name)
		for _, function := range scope.Functions {

			buf := fmt.Sprintf("%s\n", function.Summary)
			if len(function.Argument) > 0 {
				buf += fmt.Sprintf("### Arguments\n%s\n", function.Argument)
			}
			buf += fmt.Sprintf("### Example\n%s\n", function.Example)
			buf += fmt.Sprintf("\n\nGenerated On %s", time.Now().Format(time.RFC3339))
			err = ioutil.WriteFile(fmt.Sprintf("out/Perl-%s-%s.md", strings.Title(scope.Name), function.Name), []byte(buf), 0744)
			if err != nil {
				err = errors.Wrapf(err, "Failed to write file %s %s", scope.Name, function.Name)
				return
			}
		}
	}
	return
}
