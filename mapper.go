package align

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	//	"github.com/jbrukh/bayesian"
	"github.com/bbalet/stopwords"
	//"github.com/jdkato/prose/tokenize"
	"github.com/juliangruber/go-intersect"
	"github.com/labstack/echo"
	"github.com/nsip/curriculum-align/bayesian"
	"gopkg.in/fatih/set.v0"
)

type ClassifierType struct {
	Classifier *bayesian.Classifier
	Classes    []bayesian.Class
}

type Curriculum struct {
	Item         string
	Year         []string
	LearningArea string
	Strand       string
	Substrand    string
	Text         string
	Elaboration  string
	AC_Content   string `json:"AC content"`
}

func read_curriculum(directory string) ([]Curriculum, error) {
	files, _ := filepath.Glob(directory + "/*.json")
	if len(files) == 0 {
		log.Fatalln("No *.json curriculum files found in input folder" + directory)
	}
	records := make([]Curriculum, 0)
	for _, filename := range files {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("%s: ", filename)
			log.Fatalln(err)
		}
		var records1 []Curriculum
		json.Unmarshal(buf, &records1)
		records = append(records, records1...)
	}
	return records, nil
}

var classifiers map[string]ClassifierType

func normalise_tokens(tokens []string) []string {
	tokens = strings.Split(stopwords.CleanString(strings.Join(tokens, " "), "en", false), " ")
	ret := make([]string, 0)
	for _, w := range tokens {
		w = strings.TrimLeft(w, "'")
		if len(w) < 3 {
			continue
		}
		ret = append(ret, strings.ToLower(w))
	}
	return ret
}

func normalise_text(txt string) string {
	return strings.Replace(
		strings.Replace(
			strings.Replace(txt, "\u2019", "'", -1),
			"\u2018", "'", -1),
		"\u2011", "-", -1)
}

// create a classifier specific to components of the curriculum
func train_curriculum(curriculum []Curriculum, learning_area string, years []string) (ClassifierType, error) {
	sort.Slice(years, func(i, j int) bool { return years[i] > years[j] })
	key := learning_area + "_" + strings.Join(years, ",")
	if classifier, ok := classifiers[key]; ok {
		return classifier, nil
	}
	classes := make([]bayesian.Class, 0)
	class_set := set.New()
	for _, record := range curriculum {
		if len(learning_area) > 0 && record.LearningArea != learning_area {
			continue
		}
		overlap := intersect.Simple(years, record.Year)
		if len(overlap.([]interface{})) == 0 {
			continue
		}
		classes = append(classes, bayesian.Class(record.Item))
		class_set.Add(record.Item)
	}
	if len(classes) < 2 {
		return ClassifierType{}, fmt.Errorf("Not enough matching curriculum statements for classification")
	}
	classifier := bayesian.NewClassifierTfIdf(classes...)
	for _, record := range curriculum {
		if !class_set.Has(record.Item) {
			continue
		}
		train := record.Text
		if len(record.Elaboration) > 0 {
			train = train + ". " + record.Elaboration
		}
		/*
			for _, a := range Tokenise(train) {
				fmt.Printf("%s ", a)
			}
			fmt.Println()
		*/
		classifier.Learn(Tokenise("CURRIC"+record.Item, train, record), bayesian.Class(record.Item))
	}
	classifier.ConvertTermsFreqToTfIdf()
	ret := ClassifierType{Classifier: classifier, Classes: classes}
	classifiers[key] = ret // memoise
	return ret, nil
}

type AlignmentType struct {
	Item         string
	Text         string
	Score        float64
	LearningArea string
	Matches      []bayesian.MatchStruct
}

func classify_text(classif ClassifierType, curriculum_map map[string]Curriculum, input string) []AlignmentType {
	scores1, matches, _, _ := classif.Classifier.LogScores(Tokenise("", input, nil))
	response := make([]AlignmentType, 0)
	for i := 0; i < len(scores1); i++ {
		response = append(response, AlignmentType{
			Item:         string(classif.Classes[i]),
			Text:         curriculum_map[string(classif.Classes[i])].Text,
			LearningArea: curriculum_map[string(classif.Classes[i])].LearningArea,
			Score:        scores1[i],
			Matches:      matches[i]})
	}
	sort.Slice(response, func(i, j int) bool { return response[i].Score > response[j].Score })
	return response
}

var curriculum []Curriculum
var curriculum_map map[string]Curriculum

func Init() {
	var err error
	if err = InitTokeniser(); err != nil {
		log.Fatalln(err)
	}
	classifiers = make(map[string]ClassifierType)
	curriculum, err = read_curriculum("./curricula/")
	if err != nil {
		log.Fatalln(err)
	}
	curriculum_map = make(map[string]Curriculum)
	for _, record := range curriculum {
		curriculum_map[record.Item] = record
	}
}

func Align(c echo.Context) error {
	var year, learning_area, text string
	learning_area = c.QueryParam("area")
	text = c.QueryParam("text")
	year = c.QueryParam("year")
	log.Printf("Area: %s\nYears: %s\nText: %s\n", learning_area, year, text)
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
