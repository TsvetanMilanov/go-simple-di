package di

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	diTagName = "di"
)

// isFieldExported checks if the provided field is exported.
// https://golang.org/pkg/reflect/#StructField
func isFieldExported(f reflect.StructField) bool {
	return len(f.PkgPath) == 0
}

func generateDependencyMetadata(d *Dependency) *dependencyMetadata {
	vType := reflect.TypeOf(d.Value)
	value := reflect.ValueOf(d.Value)

	return &dependencyMetadata{
		Dependency:   d,
		reflectType:  vType,
		reflectValue: value,
		typeElem:     vType.Elem(),
		valueElem:    value.Elem(),
		implements:   make(map[string]bool),
	}
}

func getDependencyKey(t reflect.Type, name string) string {
	key := fmt.Sprintf("%s-%s-%s",
		t.PkgPath(),
		t.String(),
		t.Kind(),
	)
	if len(name) > 0 {
		return fmt.Sprintf("%s-%s", key, name)
	}

	return key
}

func getTags(field reflect.StructField) *diTags {
	tag, ok := field.Tag.Lookup(diTagName)
	if !ok {
		return nil
	}

	tags := strings.Split(tag, ",")
	res := &diTags{name: tags[0]}
	return res
}

func isValidValue(t reflect.Type) (isValid bool) {
	defer func() {
		if r := recover(); r != nil {
			isValid = false
		}
	}()

	kind := t.Kind()
	return kind == reflect.Ptr || kind == reflect.Interface
}
