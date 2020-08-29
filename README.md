# rdfa

RDFa is an HTML extension that help the developer to markup things and make them understandable for machines.

## how does the package work

The package works by reading the html code passed as input in the form og string, byte array or io.Reader. 
The extraction of information has some limitations compared to the RDFa protocol. 
1. The vocabularies used by the html code must be inserted at a global level, so inside the `<html>` tag at the beginning of the code. Insertion of vocabulary at a `<div>` level, or lower, are not read (yet).
2. The list of permitted vocabularies has been extracted thanks to this link https://github.com/ruby-rdf/rdf-vocab 
3. Once the list of imported vocabularies is recognised, the related attributes are identified by reading the "property" keyword. 
4. both HTML5 and XHTML are handled, take a look to the test cases. 