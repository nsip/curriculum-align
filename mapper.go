package main

import (
	"fmt"
	"log"
	//"net/http"
	"os"
	"path/filepath"
	//"runtime/pprof"
	"sort"
	"strings"
	"time"

	"github.com/jbrukh/bayesian"
	"github.com/jdkato/prose/tokenize"
	"github.com/juliangruber/go-intersect"
	"github.com/labstack/echo"
	"github.com/recursionpharma/go-csv-map"
	"gopkg.in/fatih/set.v0"
)

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

// create a classifier specific to components of the curriculum
func train_curriculum(curriculum []map[string]string, learning_area string, years []string) ([]bayesian.Class, *bayesian.Classifier) {
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
	return classes, classifier
}

type AlignmentType struct {
	Item  string
	Text  string
	Score float64
}

func classify_text(classifier *bayesian.Classifier, classes []bayesian.Class, curriculum_map map[string]string, input string) []AlignmentType {
	scores1, _, _ := classifier.LogScores(tokenize.TextToWords(input))
	response := make([]AlignmentType, 0)
	for i := 0; i < len(scores1); i++ {
		response = append(response, AlignmentType{
			Item:  string(classes[i]),
			Text:  curriculum_map[string(classes[i])],
			Score: scores1[i]})
	}
	sort.Slice(response, func(i, j int) bool { return response[i].Score > response[j].Score })
	return response
}

func main() {
	curriculum, err := read_curriculum("./curricula/")
	if err != nil {
		log.Fatalln(err)
	}
	curriculum_map := make(map[string]string)
	for _, record := range curriculum {
		curriculum_map[record["Item"]] = record["Text"]
	}

	e := echo.New()

	// TODO: memoise for efficiency?
	start := time.Now()
	classes, classifier := train_curriculum(curriculum, "Science", []string{"7", "8"})
	t := time.Now()
	log.Printf("Train curricula: %+v\n", t.Sub(start))
	start = t

	input := "I am very interested in biotechnology"

	response := classify_text(classifier, classes, curriculum_map, input)

	fmt.Printf("%+v\n", response)
	e.Logger.Fatal(e.Start(":1576"))

}
