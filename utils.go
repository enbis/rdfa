package rdfa

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func renderHtmlNode(n *html.Node) string {
	var b bytes.Buffer
	html.Render(&b, n)
	return b.String()
}

func addKeyAndValue(key string, val string, rdfBase *rdfData) {
	if _, ok := rdfBase.rdfVals[key]; !ok {
		rdfBase.rdfVals[key] = val
	}
}

func contains(s []string, e string) bool {
	for _, x := range s {
		if e == x {
			return true
		}
	}
	return false
}

func collectText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, buf)
	}
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
