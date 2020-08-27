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

type vocabolaryList struct {
	Keys []string
}

var rdfArray []string
var rdfVals map[string]string

var test = `
<html xmlns="http://www.w3.org/1999/xhtml"
xmlns:foaf="http://xmlns.com/foaf/0.1/"
xmlns:dc="http://purl.org/dc/elements/1.1/"
xhv: http://www.w3.org/1999/xhtml/vocab#
version="XHTML+RDFa 1.0" xml:lang="en">
  <head>
    <title>John's Home Page</title>
    <base href="http://example.org/john-d/" />
    <meta property="dc:creator" content="Jonathan Doe" />
    <link rel="foaf:primaryTopic" href="http://example.org/john-d/#me" />
  </head>
  <body about="http://example.org/john-d/#me">
    <h1>John's Home Page</h1>
    <p>My name is <span property="foaf:nick">John D</span> and I like
      <a href="http://www.neubauten.org/" rel="foaf:interest"
        xml:lang="de">Einstürzende Neubauten</a>.
    </p>
    <p>
      My <span rel="foaf:interest" resource="urn:ISBN:0752820907">favorite
      book is the inspiring <span about="urn:ISBN:0752820907"><cite
      property="dc:title">Weaving the Web</cite> by
      <span property="dc:creator">Tim Berners-Lee</span></span>
     </span>
    </p>
  </body>
</html>`

func main() {
	rdfVals = make(map[string]string)
	rdfArray = []string{}

	baseUri := "http://rdfa.info/"
	//baseUri := "https://www.powercms.in/blog/how-automatically-delete-docker-container-after-running-it"

	resp, err := http.Get(baseUri)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	//html = []byte(test)

	Extract(html)
}

func Extract(html []byte) {

	var vocabolary vocabolaryList
	jsonFile, err := os.Open("./rdfvocab/vocab.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	jsonByte, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(jsonByte, &vocabolary); err != nil {
		panic(err)
	}

	// extract keys vocab contained inside the <html> tag as a global val
	pattern := "(?i:(" + strings.Join(vocabolary.Keys, ")|(") + "))"
	regexec := regexp.MustCompile(pattern)
	vocabMatched := regexec.FindAllString(htmlTagSubstring(html), -1)
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
				fmt.Println("read text ", val)
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

func htmlTagSubstring(val []byte) string {
	output := ""
	in := bufio.NewReader(strings.NewReader(string(val)))
	for output == "" {
		val, err := in.ReadString('>')
		if err != nil {
			panic(err)
		}
		xval := strings.TrimSpace(val)
		if strings.HasPrefix(xval, "<html") {
			output = val
		}
	}
	fmt.Println(output)
	return output
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
		pattern := `property="` + match + `[^ ]*(\S+)(\s+)`
		regexec := regexp.MustCompile(pattern)
		regres := regexec.FindAllString(string(html), -1)
		for _, val := range regres {
			pattern = `\:(.*?)[\ ,\",\t]`
			regexec := regexp.MustCompile(pattern)
			regres = regexec.FindStringSubmatch(val)
			if len(regres) > 1 && regres[1] != "" {
				if _, ok := rdfVals[regres[1]]; !ok {
					rdfVals[regres[1]] = ""
					rdfArray = append(rdfArray, fmt.Sprintf("%s:%s", match, regres[1]))
				}
			}
		}
	}
}

func keepThisOrNext(row string, next chan string) {
	for i, rdfProp := range rdfArray {
		if strings.Contains(row, rdfProp) {
			splittedRow := strings.Split(row, "property=")[1]
			//if strings.HasSuffix(row, ">") {
			if strings.Contains(splittedRow, "content=") {
				regExecAndRemove(`content="(.*?)"`, splittedRow, rdfProp, i)
				return
			} else if strings.Contains(splittedRow, "</") {
				regExecAndRemove(`>(.*?)</`, splittedRow, rdfProp, i)
				return
			}
			//}
			for {
				newrow := <-next
				if strings.Contains(newrow, "content=") {
					regExecAndRemove(`content="(.*?)"`, newrow, rdfProp, i)
					return
				} else {
					if strings.ContainsRune(newrow, '<') {
						rdfArray = removeFromSlice(rdfArray, i)
						return
					}
					splitted := strings.Split(rdfProp, ":")[1]
					if _, ok := rdfVals[splitted]; ok {
						rdfVals[splitted] += newrow
					}
				}
			}
		}
	}
	return
}

func removeFromSlice(s []string, i int) []string {
	if len(s)-1 < i {
		panic("error length ")
	}
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
	fmt.Println("match found ", output)
	return output
}
