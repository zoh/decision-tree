package tree

import (
	"math"
	"reflect"
)

type DecisionTree struct {
	Root             *TreeItem
	CategoryAttr     string
	ignoredAttribute []string
}

type TreeItem struct {
	Match    *TreeItem
	NoMatch  *TreeItem
	Category string

	MatchedCount   int
	NoMatchedCount int
	Attribute      string
	Predicate      *Predicate
	Pivot          interface{}
}

const (
	entropyThreshold = .01
)

/*
  Predicates
*/
type Predicate func(a, b interface{}) bool

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
	tree.Root = makeTrainingTree(trainingSet, tree.CategoryAttr, tree.ignoredAttribute)
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
	Match     TrainingSet
	NoMatch   TrainingSet
	Gain      float64
	Attribute string
	Predicate *Predicate
	Pivot     interface{}
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
			if reflect.TypeOf(pivot).String() == "int" || reflect.TypeOf(pivot).String() == "float64" {
				predicate = predicateGte
			} else {
				predicate = predicateEq
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
