package tree

import (
	"fmt"
	"html/template"
	"math"
	"os"
	"reflect"
)

type DecisionTree struct {
	Root             *TreeItem
	CategoryAttr     string
	IgnoredAttribute []string
}

type TreeItem struct {
	Match    *TreeItem
	NoMatch  *TreeItem
	Category string

	MatchedCount   int
	NoMatchedCount int

	Attribute     string
	Predicate     *Predicate
	PredicateName string
	Pivot         interface{}
}

const (
	entropyThreshold = .01
)

/*
  Predicates
*/
type Predicate func(interface{}, interface{}) bool

func predicateEq(a, b interface{}) bool {
	return a == b
}
func predicateGte(a, b interface{}) bool {
	switch a.(type) {
	case float64:
		a_ := a.(float64)
		b_ := b.(float64)
		return a_ >= b_
	case int:
		a_ := a.(int)
		b_ := b.(int)
		return a_ >= b_
	}

	return false
}

/*
  Make training tree by training set.
*/
func TrainingTree(tree *DecisionTree, trainingSet TrainingSet) {
	tree.Root = makeTrainingTree(trainingSet, tree.CategoryAttr, tree.IgnoredAttribute)
}

/*
  Calc entropy of training set  by attribute.
  S = SUM(-Ni/N * log2(Ni/N))
  entropy Shanon.
*/
func entropy(set TrainingSet, attr string) float64 {
	counter := counterUniqueValues(set, attr)
	var entropy float64
	for _, val := range counter {
		b := float64(val) / float64(len(set))
		entropy += -b * math.Log(b)
	}
	return entropy
}

func counterUniqueValues(set TrainingSet, attr string) map[string]int {
	res := make(map[string]int)
	for _, item := range set {
		val := item[attr].(string)
		res[val] += 1
	}
	return res
}

/*
	Finding value of specific attribute which is most frequent
	in given array of objects.
*/
func mostFrequentValue(set TrainingSet, attr string) string {
	counter := counterUniqueValues(set, attr)
	var mostFrequentCount int
	var mostFrequentValue string

	for key, val := range counter {
		if val > mostFrequentCount {
			mostFrequentCount = val
			mostFrequentValue = key
		}
	}
	return mostFrequentValue
}

type Split struct {
	Match         TrainingSet
	NoMatch       TrainingSet
	Gain          float64
	Attribute     string
	Predicate     *Predicate
	PredicateName string
	Pivot         interface{}
}

/*
  Splitting array of objects by value of specific attribute,
  using specific predicate and pivot.
*/
func split(trainingSet TrainingSet, attr string, predicate Predicate, pivot interface{}) Split {
	var res Split
	for _, item := range trainingSet {
		if predicate(item[attr], pivot) {
			res.Match = append(res.Match, item)
		} else {
			res.NoMatch = append(res.NoMatch, item)
		}
	}
	return res
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func makeTrainingTree(trainingSet TrainingSet, categoryAttr string, ignoreAttributes []string) *TreeItem {
	initEntropy := entropy(trainingSet, categoryAttr)
	if initEntropy <= entropyThreshold {
		return &TreeItem{Category: mostFrequentValue(trainingSet, categoryAttr)}
	}
	var bestSplit Split
	// iterate all set and all attributes in item
	for _, item := range trainingSet {
		for attr, pivot := range item {
			if attr == categoryAttr || stringInSlice(attr, ignoreAttributes) {
				continue
			}

			var predicate Predicate
			var predicateName string
			if reflect.TypeOf(pivot).String() == "int" || reflect.TypeOf(pivot).String() == "float64" {
				predicate = predicateGte
				predicateName = ">="
			} else {
				predicate = predicateEq
				predicateName = "=="
			}

			// split on match/unMatch sets.
			currSplit := split(trainingSet, attr, predicate, pivot)

			// sum entropy both sets.
			matchEntropy := entropy(currSplit.Match, categoryAttr)
			noMatchEntropy := entropy(currSplit.NoMatch, categoryAttr)
			newEntropy := 0.0
			newEntropy += matchEntropy * float64(len(currSplit.Match))
			newEntropy += noMatchEntropy * float64(len(currSplit.NoMatch))
			newEntropy /= float64(len(trainingSet))

			//fmt.Println(matchEntropy, noMatchEntropy, attr, pivot, "\n")

			currGain := initEntropy - newEntropy
			if currGain > bestSplit.Gain {
				bestSplit = currSplit
				bestSplit.Gain = currGain
				bestSplit.Attribute = attr
				bestSplit.Pivot = pivot
				bestSplit.Predicate = &predicate
				bestSplit.PredicateName = predicateName
			}
		}
	}

	if bestSplit.Gain == 0 {
		// can't find optimal split
		return &TreeItem{Category: mostFrequentValue(trainingSet, categoryAttr)}
	}

	matchSubTree := makeTrainingTree(bestSplit.Match, categoryAttr, ignoreAttributes)
	notMatchSubTree := makeTrainingTree(bestSplit.NoMatch, categoryAttr, ignoreAttributes)

	return &TreeItem{
		Match:          matchSubTree,
		NoMatch:        notMatchSubTree,
		MatchedCount:   len(bestSplit.Match),
		NoMatchedCount: len(bestSplit.NoMatch),
		Attribute:      bestSplit.Attribute,
		Pivot:          bestSplit.Pivot,
		Predicate:      bestSplit.Predicate,
		PredicateName:  bestSplit.PredicateName,
	}
}

func (t DecisionTree) Predict(item TrainingItem) string {
	return predict(t.Root, item)
}

func predict(tree *TreeItem, item TrainingItem) string {
	for {
		for tree.Category != "" {
			return tree.Category
		}
		value := item[tree.Attribute]
		predicate := *tree.Predicate
		if predicate(value, tree.Pivot) {
			tree = tree.Match
		} else {
			tree = tree.NoMatch
		}
	}
}

type myWriter struct {
}

func (w myWriter) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))
	return 0, nil
}

const (
	htmlTemplate = `<html>
<head>
    <style type="text/css">
        * {
            margin: 0;
            padding: 0;
        }

        .tree ul {
            padding-top: 20px;
            position: relative;

            transition: all 0.5s;
            -webkit-transition: all 0.5s;
            -moz-transition: all 0.5s;
        }

        .tree li {
            white-space: nowrap;
            float: left;
            text-align: center;
            list-style-type: none;
            position: relative;
            padding: 20px 5px 0 5px;

            transition: all 0.5s;
            -webkit-transition: all 0.5s;
            -moz-transition: all 0.5s;
        }

        /*We will use ::before and ::after to draw the connectors*/

        .tree li::before, .tree li::after{
            content: '';
            position: absolute;
            top: 0;
            right: 50%;
            border-top: 1px solid #ccc;
            width: 50%;
            height: 20px;
        }
        .tree li::after{
            right: auto;
            left: 50%;
            border-left: 1px solid #ccc;
        }

        /*We need to remove left-right connectors from elements without
         any siblings*/
        .tree li:only-child::after, .tree li:only-child::before {
            display: none;
        }

        /*Remove space from the top of single children*/
        .tree li:only-child{
            padding-top: 0;
        }

        /*Remove left connector from first child and
         right connector from last child*/
        .tree li:first-child::before, .tree li:last-child::after{
            border: 0 none;
        }
        /*Adding back the vertical connector to the last nodes*/
        .tree li:last-child::before{
            border-right: 1px solid #ccc;
            border-radius: 0 5px 0 0;
            -webkit-border-radius: 0 5px 0 0;
            -moz-border-radius: 0 5px 0 0;
        }
        .tree li:first-child::after{
            border-radius: 5px 0 0 0;
            -webkit-border-radius: 5px 0 0 0;
            -moz-border-radius: 5px 0 0 0;
        }

        /*Time to add downward connectors from parents*/
        .tree ul ul::before{
            content: '';
            position: absolute;
            top: 0;
            left: 50%;
            border-left: 1px solid #ccc;
            width: 0;
            height: 20px;
        }

        .tree li a{
            border: 1px solid #ccc;
            padding: 5px 10px;
            text-decoration: none;
            color: #666;
            font-family: arial, verdana, tahoma;
            font-size: 11px;
            display: inline-block;

            border-radius: 5px;
            -webkit-border-radius: 5px;
            -moz-border-radius: 5px;

            transition: all 0.5s;
            -webkit-transition: all 0.5s;
            -moz-transition: all 0.5s;
        }

        /*Time for some hover effects*/
        /*We will apply the hover effect the the lineage of the element also*/
        .tree li a:hover, .tree li a:hover+ul li a {
            background: #c8e4f8;
            color: #000;
            border: 1px solid #94a0b4;
        }
        /*Connector styles on hover*/
        .tree li a:hover+ul li::after,
        .tree li a:hover+ul li::before,
        .tree li a:hover+ul::before,
        .tree li a:hover+ul ul::before{
            border-color:  #94a0b4;
        }
    </style>
</head>
<body>

<div class="tree">{{ .tree }}</div>

</body>
</html>`
)

func (t DecisionTree) SaveToHtml(out string) {
	tmpl, err := template.New("name").Parse(htmlTemplate)
	if err != nil {
		panic(err)
	}

	if out == "" {
		panic("undefined path file for save treeHtml.")
	}

	data_res := make(map[string]template.HTML)
	data_res["tree"] = template.HTML(treeToHtml(t.Root))

	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = tmpl.Execute(f, data_res)
}

func treeToHtml(tree *TreeItem) string {
	// only leafs containing category
	if tree.Category != "" {
		return `<ul>
				<li>
				<a href="#">
				<b>` + tree.Category + `</b>
				</a>
				</li>
				</ul>`
	}

	return `<ul>
		<li><a href="#">
			<b>` + tree.String() + ` ?</b>
			</a>
		<ul>
		<li>
			<a href="#">yes</a>` + treeToHtml(tree.Match) + `
		</li>
		<li>
			<a href="#">no</a>` + treeToHtml(tree.NoMatch) + `
		</li>
		</ul>
		</li></ul>`
}

func (node TreeItem) String() string {
	return fmt.Sprintf("%s %s %v", node.Attribute, node.PredicateName, node.Pivot)
}
