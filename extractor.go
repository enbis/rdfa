package rdfa

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type rdfData struct {
	rdfVals    map[string]string
	rdfArray   []string
	vocabulary vocabularyList
	editedKeys []string
}

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

	var err error

	rdfCollection := rdfData{}
	rdfCollection.rdfVals = make(map[string]string)
	rdfCollection.rdfArray = []string{}
	rdfCollection.editedKeys = []string{}

	v, err := getVocabularyType()
	if err != nil {
		return nil, err
	}
	rdfCollection.vocabulary = v

	for _, k := range rdfCollection.vocabulary.Keys {
		rdfCollection.editedKeys = append(rdfCollection.editedKeys, `:`+k)
		rdfCollection.editedKeys = append(rdfCollection.editedKeys, k+`:`)
	}

	doc, _ := html.Parse(bytes.NewReader(htmlInput))
	nodes := []*html.Node{doc}
	for len(nodes) != 0 {
		node := nodes[0]
		if node.Type == html.ElementNode && node.Data == "html" {

			var b bytes.Buffer
			html.Render(&b, node)
			htmlTag := b.String()

			if err := runVocabularyExtraction(htmlTag, &rdfCollection); err != nil {
				return nil, err
			}
		}
		for _, a := range node.Attr {
			// attributes rel, revand property are used to represent predicates
			switch a.Key {
			case "property":
				//read content or data
				preProcessNode(node, &rdfCollection, "content", a.Val)
			case "rel":
				//read href or data
				preProcessNode(node, &rdfCollection, "href", a.Val)
			case "ref":
				//read about or data
				preProcessNode(node, &rdfCollection, "about", a.Val)
			}
		}
		for n := node.FirstChild; n != nil; n = n.NextSibling {
			nodes = append(nodes, n)
		}
		nodes = nodes[1:]
	}

	res, err := json.Marshal(rdfCollection.rdfVals)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func runVocabularyExtraction(htmlTag string, rdfCollection *rdfData) error {

	pos := strings.Index(htmlTag, ">")
	if pos == -1 {
		return errors.New("html tag extraction error")
	}
	htmlTag = htmlTag[0:pos]
	pattern := "(?i:(" + strings.Join(rdfCollection.editedKeys, ")|(") + "))"
	regexec := regexp.MustCompile(pattern)
	vocabMatched := regexec.FindAllString(htmlTag, -1)
	distinctedMatches := distinctObjects(vocabMatched)

	if len(distinctedMatches) == 0 {
		return errors.New("No keys found")
	}

	rdfCollection.editedKeys = distinctedMatches

	return nil
}

func preProcessNode(node *html.Node, rdfColl *rdfData, attribute string, val string) {

	if strings.Count(val, ":") > 1 {
		splitted := strings.Split(val, " ")
		for _, s := range splitted {
			processNode(node, rdfColl, attribute, s)
		}
	} else {
		processNode(node, rdfColl, attribute, val)
	}

}

func processNode(node *html.Node, rdfColl *rdfData, attribute string, val string) {

	attributeVal := ""
	for _, a := range node.Attr {
		if a.Key == attribute {
			attributeVal = a.Val
		}
	}

	splitted := strings.Split(val, ":")
	if len(splitted) != 2 || splitted[0] == "" || splitted[1] == "" {
		return
	}
	splittedVal := splitted[0]
	splittedKey := splitted[1]

	if contains(rdfColl.rdfArray, splittedVal) {
		if attributeVal != "" {
			addKeyAndValue(splittedKey, attributeVal, rdfColl)
			return
		} else {
			text := &bytes.Buffer{}
			collectText(node, text)
			addKeyAndValue(splittedKey, text.String(), rdfColl)
			return
		}
	}
}
