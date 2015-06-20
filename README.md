# Decision tree on Golang


## Usage
```
import (
  . "github.com/zoh/decision-tree/tree"
)
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
