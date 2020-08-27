package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

type vocabolary struct {
	Keys []string
}

var rdfArray []string
var rdfVals map[string]string

func main() {
	rdfVals = make(map[string]string)
	rdfArray = []string{}

	baseUri := "http://rdfa.info/"

	resp, err := http.Get(baseUri)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var v vocabolary
	jsonFile, err := os.Open("./rdfvocab/vocab.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	jsonByte, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(jsonByte, &v); err != nil {
		panic(err)
	}

	pattern := "(?i:(" + strings.Join(v.Keys, ")|(") + "))"
	regexec := regexp.MustCompile(pattern)
	vocabMatched := regexec.FindAllString(string(html), -1)
	distinctedMatches := distinctObjects(vocabMatched)

	if len(distinctedMatches) == 0 {
		panic("no key found")
	}

	setProperty(distinctedMatches, html)

	wg := sync.WaitGroup{}
	next := make(chan string)
	eof := make(chan bool)

	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(strings.NewReader(string(html)))
		for scanner.Scan() {
			next <- scanner.Text()
		}
		close(eof)
	}()

	go func() {
		for {
			select {
			case val := <-next:
				keepThisOrNext(val, next)
			case <-eof:
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
	fmt.Printf("%v \n", rdfArray)
	fmt.Printf("%v \n", rdfVals)
}

func regExecAndRemove(pattern string, val string, prop string, toRemove int) {
	regexec := regexp.MustCompile(pattern)
	regres := regexec.FindStringSubmatch(val)
	if len(regres) > 1 && regres[1] != "" {
		splitted := strings.Split(prop, ":")[1]
		if _, ok := rdfVals[splitted]; ok {
			rdfVals[splitted] = regres[1]
			rdfArray = removeFromSlice(rdfArray, toRemove)
		}
	}
}

func setProperty(matches []string, html []byte) {
	for _, match := range matches {
		pattern := match + `[^ ]*(\S+)(\s+)`
		regexec := regexp.MustCompile(pattern)
		regres := regexec.FindAllString(string(html), -1)
		for _, val := range regres {
			pattern = `\:(.*?)[\ ,\",\t]`
			regexec := regexp.MustCompile(pattern)
			regres = regexec.FindStringSubmatch(val)
			if len(regres) > 1 && regres[1] != "" {
				if _, ok := rdfVals[regres[1]]; !ok {
					rdfVals[regres[1]] = ""
					rdfArray = append(rdfArray, fmt.Sprintf("%s%s", match, regres[1]))
				}
			}
		}
	}
}

func keepThisOrNext(row string, next chan string) {
	for i, v := range rdfArray {
		if strings.Contains(row, v) {
			if strings.HasSuffix(row, ">") {
				if strings.Contains(row, "content=") {
					regExecAndRemove(`content=(.*?)"`, row, v, i)
					return
				} else if strings.Contains(row, "</") {
					regExecAndRemove(`>(.*?)</`, row, v, i)
					return
				}
			}
			for {
				newrow := <-next
				if strings.Contains(newrow, "content=") {
					regExecAndRemove(`content="(.*?)"`, newrow, v, i)
					break
				} else {
					if strings.ContainsRune(newrow, '<') {
						rdfArray = removeFromSlice(rdfArray, i)
						break
					}
					vv := strings.Split(v, ":")[1]
					if _, ok := rdfVals[vv]; ok {
						rdfVals[vv] += newrow
					}

				}

			}
		}
	}
	return
}

func removeFromSlice(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func distinctObjects(objs []string) (distinctedObjs []string) {
	var output []string
	set := make(map[string]bool)
	for _, obj := range objs {
		if _, ok := set[obj]; !ok {
			set[obj] = true
			output = append(output, obj)
		}
	}
	return output
}
