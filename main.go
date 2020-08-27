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
	"time"
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

	//regex stuff
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

	p := "(?i:(" + strings.Join(v.Keys, ")|(") + "))"
	re := regexp.MustCompile(p)
	all := re.FindAllString(string(html), -1)
	distincted := distinctObjects(all)

	//se non ci sono elementi nell'array -> quit

	for _, single := range distincted {
		p2 := single + `[^ ]*(\S+)(\s+)`
		re2 := regexp.MustCompile(p2)
		all2 := re2.FindAllString(string(html), -1)
		for _, single2 := range all2 {
			fmt.Println(single2)
			p3 := `\:(.*?)[\ ,\",\t]`
			re3 := regexp.MustCompile(p3)
			all3 := re3.FindStringSubmatch(single2)
			if len(all3) > 1 && all3[1] != "" {
				//rdfVals[fmt.Sprintf("%s%s", single, all3[1])] = ""
				if _, ok := rdfVals[all3[1]]; !ok {
					rdfVals[all3[1]] = ""
					rdfArray = append(rdfArray, fmt.Sprintf("%s%s", single, all3[1]))
				}
			}
			fmt.Println(all3)
		}
	}

	wg := sync.WaitGroup{}
	next := make(chan string)
	eof := make(chan bool)

	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(strings.NewReader(string(html)))
		for scanner.Scan() {
			time.Sleep(100 * time.Millisecond)
			next <- scanner.Text()
		}
		close(eof)
	}()

	go func() {
		for {
			select {
			case val := <-next:
				fmt.Println("val read: ", val)
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

func keepThisOrNext(row string, next chan string) {
	for i, v := range rdfArray {
		if strings.Contains(row, v) {
			if strings.HasSuffix(row, ">") {
				if strings.Contains(row, "content=") {
					fmt.Println("Estrarre content ", row)
					p3 := `content=(.*?)"`
					re3 := regexp.MustCompile(p3)
					all3 := re3.FindStringSubmatch(row)
					if len(all3) > 1 && all3[1] != "" {
						vv := strings.Split(v, ":")[1]
						if _, ok := rdfVals[vv]; ok {
							rdfVals[vv] = all3[1]
							rdfArray = removeFromSlice(rdfArray, i)
							return
						}
					}
					return
				} else if strings.Contains(row, "</") {
					fmt.Println("Estrarre regex ", row)
					p3 := `>(.*?)</`
					re3 := regexp.MustCompile(p3)
					all3 := re3.FindStringSubmatch(row)
					if len(all3) > 1 && all3[1] != "" {
						vv := strings.Split(v, ":")[1]
						if _, ok := rdfVals[vv]; ok {
							rdfVals[vv] = all3[1]
							rdfArray = removeFromSlice(rdfArray, i)
							return
						}
					}
					return
				}
			}
			fmt.Println("Estrarre next row ", row)
			for {
				newrow := <-next
				if strings.Contains(newrow, "content=") {
					fmt.Println("Estrarre content ", newrow)
					p3 := `content="(.*?)"`
					re3 := regexp.MustCompile(p3)
					all3 := re3.FindStringSubmatch(newrow)
					if len(all3) > 1 && all3[1] != "" {
						vv := strings.Split(v, ":")[1]
						if _, ok := rdfVals[vv]; ok {
							rdfVals[vv] = all3[1]
							rdfArray = removeFromSlice(rdfArray, i)
						}
					}
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
			fmt.Println("Estrarre next row ", row)
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
