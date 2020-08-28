package rdfa

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type vocabolaryList struct {
	Keys []string
}

func getVocabolaryType() (vocabolaryList, error) {
	var vocabolary vocabolaryList
	jsonFile, err := os.Open("./rdfvocab/vocab.json")
	if err != nil {
		return vocabolary, err
	}
	defer jsonFile.Close()
	jsonByte, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return vocabolary, err
	}

	if err = json.Unmarshal(jsonByte, &vocabolary); err != nil {
		return vocabolary, err
	}

	return vocabolary, nil
}
