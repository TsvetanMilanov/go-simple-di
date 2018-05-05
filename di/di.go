package di

import (
	"errors"
	"fmt"
	"reflect"
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
		if !isValidValue(dType) {
			return fmt.Errorf("%s should be pointer or interface", dType.String())
		}

		meta := generateDependencyMetadata(d)
		key := getDependencyKey(meta.reflectType, d.Name)
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
	if !isValidValue(resType) {
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

	// Mark as complete to avoid circular dependency recursion.
	// This requires each return with error to set the complete property to false.
	d.complete = true
	for i := 0; i < d.typeElem.NumField(); i++ {
		field := d.typeElem.Field(i)
		tags, err := getTags(field)
		if err != nil {
			d.complete = false
			return fmt.Errorf("[%s] %s", d.reflectType.String(), err.Error())
		}

		if tags == nil {
			continue
		}

		if !isValidValue(field.Type) || !isFieldExported(field) {
			d.complete = false
			return fmt.Errorf("[%s] cannot set field %s", d.reflectType.String(), field.Name)
		}

		fieldDep := c.findDependency(field.Type, tags.name)
		if fieldDep == nil {
			d.complete = false
			return fmt.Errorf("[%s] unable to find registered dependency: %s", d.reflectType.String(), field.Name)
		}

		err = c.resolveCore(fieldDep)
		if err != nil {
			d.complete = false
			return fmt.Errorf("[%s] %s", d.reflectType.String(), err.Error())
		}

		d.valueElem.Field(i).Set(fieldDep.reflectValue)
	}

	return nil
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
		key := getDependencyKey(t, name)
		return c.dependencies[key]
	}

	return nil
}
