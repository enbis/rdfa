package rdfa

import (
	"bytes"
	"net/http"
	"testing"
)

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

func TestExtractor(t *testing.T) {
	var err error
	baseUri := "http://rdfa.info/"

	resp, err := http.Get(baseUri)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = Extract(resp.Body)
	if err != nil {
		t.Error(err)
	}

	var a []byte
	a, err = Extract([]byte(test))
	if err != nil {
		t.Error(err)
	}

	var b []byte
	b, err = Extract(test)
	if err != nil {
		t.Error(err)
	}

	x := bytes.Compare(a, b)
	if x != 0 {
		t.Error("same input different responses")
	}

	expected := "input value type not allowed"
	_, err = Extract(1)
	if err.Error() != expected {
		t.Error("error")
	}

}
