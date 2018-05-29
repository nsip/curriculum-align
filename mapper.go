package align

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jbrukh/bayesian"
	"github.com/jdkato/prose/tokenize"
	"github.com/juliangruber/go-intersect"
	"github.com/labstack/echo"
	"github.com/recursionpharma/go-csv-map"
	"gopkg.in/fatih/set.v0"
)

type ClassifierType struct {
	Classifier *bayesian.Classifier
	Classes    []bayesian.Class
}

// assumes tab-delimited file with header.
// Expects to find fields Item (identifier), Year, LearningArea, Text, and optionally Elaborations
// Year can contain multiple values; it is ";"-delimited
func read_curriculum(directory string) ([]map[string]string, error) {
	files, _ := filepath.Glob(directory + "/*.txt")
	if len(files) == 0 {
		log.Fatalln("No *.txt curriculum files found in input folder" + directory)
	}
	records := make([]map[string]string, 0)
	for _, filename := range files {
		buf, err := os.Open(filename)
		if err != nil {
			log.Printf("%s: ", filename)
			log.Fatalln(err)
		}
		defer buf.Close()
		reader := csvmap.NewReader(buf)
		reader.Reader.Comma = '\t'
		columns, err := reader.ReadHeader()
		if err != nil {
			log.Printf("%s: ", filename)
			log.Fatalln(err)
		}
		reader.Columns = columns
		records1, err := reader.ReadAll()
		if err != nil {
			log.Printf("%s: ", filename)
			log.Fatalln(err)
		}
		records = append(records, records1...)
	}
	return records, nil
}

var classifiers map[string]ClassifierType

// create a classifier specific to components of the curriculum
func train_curriculum(curriculum []map[string]string, learning_area string, years []string) (ClassifierType, error) {
	sort.Slice(years, func(i, j int) bool { return years[i] > years[j] })
	key := learning_area + "_" + strings.Join(years, ",")
	if classifier, ok := classifiers[key]; ok {
		return classifier, nil
	}
	classes := make([]bayesian.Class, 0)
	class_set := set.New()
	for _, record := range curriculum {
		if record["LearningArea"] != learning_area {
			continue
		}
		overlap := intersect.Simple(years, strings.Split(strings.Replace(record["Year"], "\"", "", -1), ";"))
		if len(overlap.([]interface{})) == 0 {
			continue
		}
		classes = append(classes, bayesian.Class(record["Item"]))
		class_set.Add(record["Item"])
	}
	if len(classes) < 2 {
		return ClassifierType{}, fmt.Errorf("Not enough matching curriculum statements for classification")
	}
	classifier := bayesian.NewClassifierTfIdf(classes...)
	for _, record := range curriculum {
		if !class_set.Has(record["Item"]) {
			continue
		}
		train := record["Text"]
		if train2, ok := record["Elaborations"]; ok {
			train = train + ". " + train2
		}
		classifier.Learn(tokenize.TextToWords(train), bayesian.Class(record["Item"]))
	}
	classifier.ConvertTermsFreqToTfIdf()
	ret := ClassifierType{Classifier: classifier, Classes: classes}
	classifiers[key] = ret // memoise
	return ret, nil
}

type AlignmentType struct {
	Item  string
	Text  string
	Score float64
}

func classify_text(classif ClassifierType, curriculum_map map[string]string, input string) []AlignmentType {
	scores1, _, _ := classif.Classifier.LogScores(tokenize.TextToWords(input))
	response := make([]AlignmentType, 0)
	for i := 0; i < len(scores1); i++ {
		response = append(response, AlignmentType{
			Item:  string(classif.Classes[i]),
			Text:  curriculum_map[string(classif.Classes[i])],
			Score: scores1[i]})
	}
	sort.Slice(response, func(i, j int) bool { return response[i].Score > response[j].Score })
	return response
}

var curriculum []map[string]string
var curriculum_map map[string]string

func Init() {
	var err error
	classifiers = make(map[string]ClassifierType)
	curriculum, err = read_curriculum("./curricula/")
	if err != nil {
		log.Fatalln(err)
	}
	curriculum_map = make(map[string]string)
	for _, record := range curriculum {
		curriculum_map[record["Item"]] = record["Text"]
	}
}

func Align(c echo.Context) error {
	var year, learning_area, text string
	learning_area = c.QueryParam("area")
	text = c.QueryParam("text")
	year = c.QueryParam("year")
	log.Printf("Area: %s\nYears: %s\nText: %s\n", learning_area, year, text)
	if learning_area == "" {
		err := fmt.Errorf("area parameter not supplied")
		c.String(http.StatusBadRequest, err.Error())
		return err
	}
	if text == "" {
		err := fmt.Errorf("text parameter not supplied")
		c.String(http.StatusBadRequest, err.Error())
		return err
	}
	if year == "" {
		year = "K,P,1,2,3,4,5,6,7,8,9,10,11,12"
	}
	classifier, err := train_curriculum(curriculum, learning_area, strings.Split(year, ","))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return err
	}
	response := classify_text(classifier, curriculum_map, text)
	return c.JSONPretty(http.StatusOK, response, "  ")
}
