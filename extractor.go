package rdfa

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

var rdfVals (map[string]string)
var rdfArray []string

func Extract(i interface{}) ([]byte, error) {
	var res []byte
	var err error
	switch reflect.ValueOf(i).Interface().(type) {
	case []byte:
		b, ok := i.([]byte)
		if !ok {
			err = errors.New("unable to execute interface conversion")
			break
		}
		res, err = runExtraction(b)
	case string:
		s, ok := i.(string)
		if !ok {
			err = errors.New("unable to execute interface conversion")
			break
		}
		res, err = runExtraction([]byte(s))
	case io.Reader:
		ioR, ok := i.(io.Reader)
		if !ok {
			err = errors.New("unable to execute interface conversion")
			break
		}
		r, errR := ioutil.ReadAll(ioR)
		if errR != nil {
			err = errR
			break
		}
		res, err = runExtraction(r)
	default:
		err = errors.New("input value type not allowed")
	}
	return res, err
}

func runExtraction(html []byte) ([]byte, error) {

	rdfVals = make(map[string]string)
	rdfArray = []string{}

	vocabolary, err := getVocabolaryType()
	if err != nil {
		return nil, err
	}
	editedKeys := []string{}
	for _, k := range vocabolary.Keys {
		editedKeys = append(editedKeys, `:`+k)
		editedKeys = append(editedKeys, k+`:`)
	}
	// extract keys vocab contained inside the <html> tag as a global val
	pattern := "(?i:(" + strings.Join(editedKeys, ")|(") + "))"
	regexec := regexp.MustCompile(pattern)
	vocabMatched := regexec.FindAllString(htmlTagSubstring(html), -1)
	distinctedMatches := distinctObjects(vocabMatched)

	if len(distinctedMatches) == 0 {
		return nil, errors.New("no rdfa keys found inside the first html tag")
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

	res, err := json.Marshal(rdfVals)

	if err != nil {
		return nil, err
	}

	return res, nil
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
			output = xval
		}
	}
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
			output = append(output, strings.Trim(obj, ":"))
		}
	}
	return output
}
