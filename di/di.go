package di

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Dependency is di dependency.
type Dependency struct {
	Name  string
	Value interface{}
}

// NewContainer creates new di container.
func NewContainer() *Container {
	return &Container{dependencies: make(map[string]*dependencyMetadata)}
}

// Container is the di container.
type Container struct {
	dependencies map[string]*dependencyMetadata
}

// Register adds the provided dependencies to the container.
func (c *Container) Register(deps ...*Dependency) error {
	for _, d := range deps {
		dType := reflect.TypeOf(d.Value)
		if !c.isValidValue(dType) {
			return fmt.Errorf("%s should be pointer or interface", dType.String())
		}

		meta := c.generateDependencyMetadata(d)
		key := c.getDependencyKey(meta.reflectType, d.Name)
		if _, ok := c.dependencies[key]; ok {
			return fmt.Errorf("duplicate dependency: %s", key)
		}

		c.dependencies[key] = meta
	}

	return nil
}

// ResolveAll pupulates the marked dependencies with the registered
// dependencies.
func (c *Container) ResolveAll() error {
	for _, d := range c.dependencies {
		err := c.resolveCore(d)
		if err != nil {
			return err
		}
	}

	return nil
}

// Resolve sets the out parameter to the resolved dependency value.
func (c *Container) Resolve(out interface{}) error {
	return c.ResolveByName("", out)
}

// ResolveByName sets the out parameter to the resolved by name dependency value.
func (c *Container) ResolveByName(name string, out interface{}) error {
	resType := reflect.TypeOf(out)
	if !c.isValidValue(resType) {
		return errors.New("the out parameter must be a pointer")
	}

	isInterface := resType.Elem().Kind() == reflect.Interface
	var dep *dependencyMetadata
	if isInterface {
		// Currently (Go 1.9.4) the reflect.TypeOf will return nil
		// the it is called with empty interface value -> https://golang.org/pkg/reflect/#TypeOf
		// For example:
		//   var a someInterface
		//   fmt.Println(reflect.TypeOf(a))
		// That's why we need to work with pointer to interface in this method.
		// The result reflect.Type of the Elem() method executed on pointer to interface gives
		// the correct type and we can work with it.
		dep = c.findDependency(resType.Elem(), name)
	} else {
		dep = c.findDependency(resType, name)
	}

	if dep == nil {
		return fmt.Errorf("unable to find registered dependency: %s", resType.String())
	}

	err := c.resolveCore(dep)
	if err != nil {
		return err
	}

	var resValue reflect.Value
	if isInterface {
		// We need to use the actual pointer reflect value to set it
		// to the provided interface.
		resValue = dep.reflectValue
	} else {
		resValue = dep.valueElem
	}

	reflect.ValueOf(out).Elem().Set(resValue)
	return nil
}

func (c *Container) resolveCore(d *dependencyMetadata) error {
	if d.complete {
		return nil
	}

	for i := 0; i < d.typeElem.NumField(); i++ {
		field := d.typeElem.Field(i)
		tags := c.getTags(field)
		if tags == nil {
			continue
		}

		if !c.isValidValue(field.Type) {
			return fmt.Errorf("[%s] cannot set field %s", d.reflectType.String(), field.Name)
		}

		fieldDep := c.findDependency(field.Type, tags.name)
		if fieldDep == nil {
			return fmt.Errorf("[%s] unable to find registered dependency: %s", d.reflectType.String(), field.Name)
		}

		err := c.resolveCore(fieldDep)
		if err != nil {
			return fmt.Errorf("[%s] %s", d.reflectType.String(), err.Error())
		}

		d.valueElem.Field(i).Set(fieldDep.reflectValue)
	}

	d.complete = true
	return nil
}

func (c *Container) getDependencyKey(t reflect.Type, name string) string {
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

func (c *Container) getTags(field reflect.StructField) *diTags {
	tag, ok := field.Tag.Lookup(diTagName)
	if !ok {
		return nil
	}

	tags := strings.Split(tag, ",")
	res := &diTags{}
	if len(tags) == 0 {
		return res
	}

	res.name = tags[0]

	return res
}

func (c *Container) isValidValue(t reflect.Type) (isValid bool) {
	defer func() {
		if r := recover(); r != nil {
			isValid = false
		}
	}()

	kind := t.Kind()
	return kind == reflect.Ptr || kind == reflect.Interface
}

func (c *Container) findDependency(t reflect.Type, name string) *dependencyMetadata {
	if t.Kind() == reflect.Interface {
		for _, v := range c.dependencies {
			if len(name) > 0 && v.Name != name {
				// Skip other checks if name is provided and it does not match.
				continue
			}

			if v.implements[t.String()] || v.reflectType.Implements(t) {
				v.implements[t.String()] = true
				return v
			}
		}
	} else {
		key := c.getDependencyKey(t, name)
		return c.dependencies[key]
	}

	return nil
}

func (c *Container) generateDependencyMetadata(d *Dependency) *dependencyMetadata {
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

const (
	diTagName = "di"
)

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
