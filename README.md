# curriculum-align
Web service to align free text to curriculum standards as document classification

NOTE: This is experimental and proof-of-concept code

This code builds on https://github.com/nsip/curriculum-mapper, putting in place a 
document classifier (https://en.wikipedia.org/wiki/Tf–idf) to classify arbitrary
text as aligning to the curriculum items the code is provisioned with,
and outputting the alignments as a web service.

The distribution creates the web service as a library (`Align()`), and the `cmd` directory contains a sample shell for it. In the sample shell, the web service runs on port 1576.

The web service takes the following arguments:

````
GET http://localhost:1576/align?yr=X,Y,Z&area=W&text=....
````
where _yr_ is the year level (and can be comma-delimited), _area_ is the learning area, and _text_ is the text to be aligned. The _area_ and _text_ parameters are obligatory. For example:

````
http://localhost:1576/align?area=Science&year=6,7,8&text=Biotechnology
````

The response is a JSON list of structs, one for each curriculum standard that the service is configured for, with the following fields:

* Item: the identifier of the curriculum item whose alignment is reported
* Text: the text of the curriculum item whose alignment is reported
* Score: the score of the alignment. This is the score generated by github.com/jbrukh/bayesian: it is a negative number, and the higher the number (i.e. the closer to zero), the better the alignment of the text to the curriculum standard.

To use embedded in other labstack.echo webservers, replicate the `cmd/main.go` main() code:

````
align.Init()
e := echo.New()
e.GET("/align", align.Align)
````

The curricula that the web service is configured to read are JSON files in the `curricula` folder of the executable; in this distribution, sample files are in `cmd/curricula`.
The JSON structure required is an array of JSON objects, with the following keys:

* Item (string): the curriculum standard code.
* Year ([]string): the year levels to which the curriculum standard applies. (NSW Syllabus stages are to be translated to year levels).
* LearningArea (string): the learning area of the curriculum standard
* Strand (string): the strand of the curriculum standard
* Substrand (string): the substrand of the curriculum standard
* Text (string): the text of the curriculum standard
* Elaboration (string): the elaboration text of the curriculum standard

## 1576

https://en.wikipedia.org/wiki/Curriculum:

> The word "curriculum" began as a Latin word which means "a race" or "the course of a race" (which in turn derives from the verb _currere_ meaning "to run/to proceed"). The first known use in an educational context is in the [_Professio Regia_](https://books.google.com.au/books?id=bG5EAAAAcAAJ&printsec=frontcover&hl=el&source=gbs_ge_summary_r&cad=0#v=onepage&q=curriculum&f=false), a work by University of Paris professor [Petrus Ramus](https://en.wikipedia.org/wiki/Petrus_Ramus) published [posthumously](https://en.wikipedia.org/wiki/St._Bartholomew%27s_Day_massacre) in 1576.
