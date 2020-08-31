package rdfa

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/net/html"
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

func runExtraction(htmlInput []byte) ([]byte, error) {

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
	vocabMatched := regexec.FindAllString(htmlTagSubstring(htmlInput), -1)
	distinctedMatches := distinctObjects(vocabMatched)

	if len(distinctedMatches) == 0 {
		return nil, errors.New("no rdfa keys found inside the first html tag")
	}

	setProperty(distinctedMatches, htmlInput)

	wg := sync.WaitGroup{}
	next := make(chan *html.Node)
	eof := make(chan bool)

	wg.Add(2)

	go func() {
		defer wg.Done()

		doc, _ := html.Parse(bytes.NewReader(htmlInput))
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode {
				for _, a := range n.Attr {
					if a.Key == "property" {
						next <- n
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)

		close(eof)
	}()

	go func() {
		for {
			select {
			case val := <-next:
				processNode(val)
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

func renderHtmlNode(n *html.Node) string {
	var b bytes.Buffer
	html.Render(&b, n)
	return b.String()
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

func setAndRemove(val string, prop string, toRemove int) {

	splitted := strings.Split(prop, ":")[1]
	if _, ok := rdfVals[splitted]; ok {
		rdfVals[splitted] = val
		rdfArray = removeFromSlice(rdfArray, toRemove)
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

func processNode(node *html.Node) {

	row := renderHtmlNode(node)
	content := ""
	for _, a := range node.Attr {
		if a.Key == "content" {
			content = a.Val
		}
	}

	for i, rdfProp := range rdfArray {
		if strings.Contains(row, rdfProp) {
			if content != "" {
				setAndRemove(content, rdfProp, i)
				return
			} else {
				text := &bytes.Buffer{}
				collectText(node, text)
				setAndRemove(text.String(), rdfProp, i)
				return
			}
		}
	}
}

func collectText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, buf)
	}
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
