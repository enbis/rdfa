package rdfa

import (
	"encoding/json"
)

type vocabolaryList struct {
	Keys []string
}

func getVocabolaryType() (vocabolaryList, error) {
	var vocabolary vocabolaryList

	if err := json.Unmarshal([]byte(jsonFile), &vocabolary); err != nil {
		return vocabolary, err
	}

	return vocabolary, nil
}

var jsonFile = `{
    "keys" : [
        "acl",
        "as",
        "bf2",
        "bibo",
        "CERT", 
        "CNT", 
        "DataCite", 
        "DBO",  
        "DC",  
        "DC11",  
        "DCAT",  
        "DCMIType",  
        "DISCO",  
        "DOAP",  
        "DWC",  
        "EARL",  
        "EBUCore",  
        "EDM",
        "EXIF",  
        "Fcrepo4",  
        "FOAF",  
        "GEO",  
        "GEOJSON",  
        "GEONAMES",  
        "GR",  
        "GS1",  
        "HT",  
        "HYDRA",  
        "IANA",  
        "ICAL",  
        "Identifiers",  
        "IIIF",  
        "JSONLD",  
        "LDP",  
        "LRMI",  
        "MA",  
        "MADS", 
        "MARCRelators",  
        "MO",  
        "MODS",  
        "NFO",  
        "OA",
        "OG", 
        "OGC", 
        "ORE",  
        "ORG", 
        "PCDM",  
        "PPLAN",  
        "PREMIS", 
        "PremisEventType",  
        "PROV", 
        "PTR",  
        "RightsStatements",  
        "RSA",  
        "RSS",
        "SCHEMA",  
        "SD",  
        "SH", 
        "SIOC",  
        "SiocServices", 
        "SiocTypes", 
        "SKOS",  
        "SKOSXL",  
        "V", 
        "VMD",  
        "VCARD",  
        "VMD",  
        "VOID",  
        "VS",  
        "WDRS",  
        "WOT",  
        "XKOS",  
        "XHTML",  
        "XHV"
    ]
}`
