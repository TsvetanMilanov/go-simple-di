package di_test

import (
	"fmt"

	"github.com/TsvetanMilanov/go-simple-di/di"
)

// struct with named and unnamed dependencies.
type root struct {
	Nested *nested `di:""`
	Named  *named  `di:"name=someName"`
}

// struct with interface dependency.
type nested struct {
	W worker `di:""`
}

type named struct {
	name string
}

type worker interface {
	Work() string
}

type builder struct {
	work string
}

func (b *builder) Work() string {
	return b.work
}

func Example() {
	// create the dependencies.
	r := &root{}
	nam := &named{name: "named"}
	nes := &nested{}
	b := &builder{work: "Build"}

	// create the di container and add all dependencies to it.
	c := di.NewContainer()
	err := c.Register(
		&di.Dependency{Value: r},
		&di.Dependency{Value: nam, Name: "someName"}, // register the named dependency with the same name as in the struct definition.
		&di.Dependency{Value: nes},
		&di.Dependency{Value: b},
	)
	if err != nil {
		panic(err)
	}

	// resolve all registered dependencies
	err = c.ResolveAll()
	if err != nil {
		panic(err)
	}

	// use the resolved dependencies
	fmt.Println("Worker: ", r.Nested.W.Work())
	fmt.Println("Named: ", r.Named.name)
}
