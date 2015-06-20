# Decision tree on Golang


## Usage
```
tree := DecisionTree{CategoryAttr: "param", ignoredAttribute: []string{}}
TrainingTree(&tree, trainingSet)

//...
// category item
var item TrainingItem
category := tree.Predict(item)
```

## Test
```
go test
```