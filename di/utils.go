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

func getTags(field reflect.StructField) (*diTags, error) {
	tag, ok := field.Tag.Lookup(diTagName)
	if !ok {
		return nil, nil
	}

	tags := strings.Split(tag, ",")
	res := new(diTags)
	// Check for empty tags.
	if len(tags) == 1 && len(tags[0]) == 0 {
		// Return default values.
		return res, nil
	}

	for _, tag := range tags {
		tagContent := strings.Split(tag, "=")
		if len(tagContent) != 2 {
			return nil, getInvalidTagErr(tag)
		}

		k := tagContent[0]
		v := tagContent[1]
		if len(v) == 0 {
			return nil, getInvalidTagErr(tag)
		}

		switch k {
		case "name":
			res.name = v
		default:
			return nil, getInvalidTagErr(tag)
		}
	}

	return res, nil
}

func getInvalidTagErr(tag string) error {
	return fmt.Errorf("invalid tag configuration '%s', expecting <key>=<value>", tag)
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

func isPointerTypePointerToInterface(t reflect.Type) bool {
	return t.Elem().Kind() == reflect.Interface
}
