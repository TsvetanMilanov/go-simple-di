package di

import "reflect"

type diTags struct {
	name string
}

type dependencyMetadata struct {
	*Dependency
	reflectType  reflect.Type
	reflectValue reflect.Value
	complete     bool
	typeElem     reflect.Type
	valueElem    reflect.Value
	implements   map[string]bool
}
