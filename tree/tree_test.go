package tree

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCounterUnique(t *testing.T) {
	trainingSet := TrainingSet{
		TrainingItem{"param": "yes"},
		TrainingItem{"param": "no"},
		TrainingItem{"param": "yes"},
		TrainingItem{"param": "yes"},
	}
	values := counterUniqueValues(trainingSet, "param")
	need := make(map[string]int, 2)
	need["yes"] = 3
	need["no"] = 1
	assert.Equal(t, values, need, "значения должны быть эквивалентны")
}

func TestEntropy(t *testing.T) {
	trainingSet := TrainingSet{
		TrainingItem{"param": "yes"},
		TrainingItem{"param": "no"},
		TrainingItem{"param": "yes"},
		TrainingItem{"param": "yes"},
	}
	val := entropy(trainingSet, "param")
	assert.Equal(t, val > 0, true)
}

func TestSplit(t *testing.T) {
	trainingSet := TrainingSet{
		TrainingItem{"param": "yes", "age": 20},
		TrainingItem{"param": "no", "age": 30},
		TrainingItem{"param": "yes", "age": 1},
	}
	split_ := split(trainingSet, "age", predicateGte, 20)
	assertSplit := Split{
		Match:   TrainingSet{trainingSet[0], trainingSet[1]},
		NoMatch: TrainingSet{trainingSet[2]},
	}
	assert.EqualValues(t, split_, assertSplit)
}

func TestMakeTrainingTree(t *testing.T) {
	trainingSet := TrainingSet{
		TrainingItem{"person": "Homer", "hairLength": 0, "weight": 250, "age": 36, "sex": "male"},
		TrainingItem{"person": "Marge", "hairLength": 10, "weight": 150, "age": 34, "sex": "female"},
		TrainingItem{"person": "Bart", "hairLength": 2, "weight": 90, "age": 10, "sex": "male"},
		TrainingItem{"person": "Lisa", "hairLength": 6, "weight": 78, "age": 8, "sex": "female"},
		TrainingItem{"person": "Maggie", "hairLength": 4, "weight": 20, "age": 1, "sex": "female"},
		TrainingItem{"person": "Abe", "hairLength": 1, "weight": 170, "age": 70, "sex": "male"},
		TrainingItem{"person": "Selma", "hairLength": 8, "weight": 160, "age": 41, "sex": "female"},
		TrainingItem{"person": "Otto", "hairLength": 10, "weight": 180, "age": 38, "sex": "male"},
		TrainingItem{"person": "Krusty", "hairLength": 6, "weight": 200, "age": 45, "sex": "male"}}
	rootItem := makeTrainingTree(trainingSet, "sex", []string{"person"})

	item := TrainingItem{
		"person": "Comic guy", "hairLength": 8, "weight": 290, "age": 38}
	assert.Equal(t, predict(rootItem, item), "male")
}
