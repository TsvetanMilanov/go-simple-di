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

// ResolveAll populates the marked dependencies with the registered
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
	return c.resolveWithFinder(func(isInterface bool) *dependencyMetadata {
		return c.findDependency(out, name)
	}, out)
}

// ResolveNew returns new instance of the provided type.
// The dependencies of the instance marked for resolving will not be new.
func (c *Container) ResolveNew(out interface{}) error {
	return c.resolveWithFinder(func(isInterface bool) *dependencyMetadata {
		dep := c.findDependency(out, "")
		var resTypeElem reflect.Type
		if dep == nil {
			if isInterface {
				// No dependency which implements the interface was registered.
				return nil
			}

			// The out is struct which is not registered in the container.
			resTypeElem = reflect.TypeOf(out).Elem()
		} else {
			// The out is struct which is registered in the container.
			resTypeElem = dep.typeElem
		}

		return generateDependencyMetadata(&Dependency{Value: reflect.New(resTypeElem).Interface()})
	}, out)
}

func (c *Container) resolveWithFinder(finder func(isInterface bool) *dependencyMetadata, out interface{}) error {
	resType := reflect.TypeOf(out)
	if !isValidValue(resType) {
		return errors.New("the out parameter must be a pointer")
	}

	isInterface := isPointerTypePointerToInterface(resType)
	dep := finder(isInterface)
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

func (c *Container) findDependency(out interface{}, name string) *dependencyMetadata {
	outType := reflect.TypeOf(out)
	var dep *dependencyMetadata
	if isPointerTypePointerToInterface(outType) {
		// Currently (Go 1.9.4) the reflect.TypeOf will return nil
		// the it is called with empty interface value -> https://golang.org/pkg/reflect/#TypeOf
		// For example:
		//   var a someInterface
		//   fmt.Println(reflect.TypeOf(a))
		// That's why we need to work with pointer to interface in this method.
		// The result reflect.Type of the Elem() method executed on pointer to interface gives
		// the correct type and we can work with it.
		dep = c.findDependencyCore(outType.Elem(), name)
	} else {
		dep = c.findDependencyCore(outType, name)
	}

	return dep
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

		fieldDep := c.findDependencyCore(field.Type, tags.name)
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

func (c *Container) findDependencyCore(t reflect.Type, name string) *dependencyMetadata {
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
