# rdfa

RDFa is an HTML extension that help the developer to markup things and make them understandable for machines.

## how does the package work

The package works by reading the html code passed as input in the form of string, byte array or io.Reader. 
The extraction of information has some limitations compared to the RDFa protocol. 
1. The vocabularies used by the html code must be inserted at a global level, inside the `<html>` tag at the beginning of the code.
2. The list of permitted vocabularies has been extracted thanks to this link https://github.com/ruby-rdf/rdf-vocab  
3. both HTML5 and XHTML are handled, take a look to the test cases. 

## changelog

### v0.0.5
RDFa algorithm bug fixed

### v0.0.4 
Improved the data extraction process, now based on tag by tag html code's reading instead of line by line string.