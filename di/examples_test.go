package di_test

import (
	"fmt"

	"github.com/TsvetanMilanov/go-simple-di/di"
)

func ExampleContainer_Register() {
	type p struct {
		prop string
	}
	type v struct {
		value int
		Prop  *p `di:""`
	}
	type d1 struct {
		Dep  *v `di:""`
		Prop *p `di:"name=resolveMe"`
	}

	c := di.NewContainer()
	err := c.Register(
		&di.Dependency{Value: new(d1)},
		&di.Dependency{Value: &v{value: 5}},
		&di.Dependency{Value: &p{prop: "named"}, Name: "resolveMe"}, // Register named dependency.
		&di.Dependency{Value: &p{prop: "unnamed"}},                  // Register unnamed dependency.
	)
	if err != nil {
		panic(err)
	}

	res := new(d1)
	c.Resolve(res)

	fmt.Println("Dep:", res.Dep.value)
	fmt.Println("Prop named:", res.Prop.prop)
	fmt.Println("Prop unnamed:", res.Dep.Prop.prop)
	// Output:
	// Dep: 5
	// Prop named: named
	// Prop unnamed: unnamed
}

func ExampleContainer_ResolveByName() {
	type dep struct {
		value string
	}

	c := di.NewContainer()
	c.Register(
		&di.Dependency{Value: &dep{value: "unnamed"}},
		&di.Dependency{Value: &dep{value: "named"}, Name: "pickMe"},
	)

	named := new(dep)
	unnamed := new(dep)
	err := c.ResolveByName("pickMe", named)
	if err != nil {
		panic(err)
	}

	fmt.Println("Named:", named.value)

	c.Resolve(unnamed)
	fmt.Println("Unnamed:", unnamed.value)
	// Output:
	// Named: named
	// Unnamed: unnamed
}

func ExampleContainer_ResolveNew_registered() {
	type dep struct {
		value int
	}

	c := di.NewContainer()
	c.Register(
		&di.Dependency{Value: &dep{value: 400}},
	)

	res := new(dep)
	err := c.ResolveNew(res)
	if err != nil {
		panic(err)
	}

	fmt.Println("Result:", res.value) // The result won't be 400.
	// Output:
	// Result: 0
}

func ExampleContainer_ResolveNew_notRegistered() {
	type dep struct {
		value int
	}

	c := di.NewContainer()
	res := new(dep)
	err := c.ResolveNew(res)
	if err != nil {
		panic(err)
	}

	fmt.Println("Result:", res.value)
	// Output:
	// Result: 0
}

func ExampleContainer_ResolveAll() {
	type d1 struct {
		v string
	}
	// First hierarchy.
	type h1 struct {
		Dep *d1 `di:""`
	}

	type d2 struct {
		v string
	}
	// Second hierarchy.
	type h2 struct {
		Dep *d2 `di:""`
	}

	c := di.NewContainer()
	first := new(h1)
	second := new(h2)
	c.Register(
		&di.Dependency{Value: &d2{v: "d2"}},
		&di.Dependency{Value: &d1{v: "d1"}},
		&di.Dependency{Value: first},
		&di.Dependency{Value: second},
	)

	err := c.ResolveAll()
	if err != nil {
		panic(err)
	}

	fmt.Println("First:", first.Dep.v)
	fmt.Println("Second:", second.Dep.v)
	// Output:
	// First: d1
	// Second: d2
}

func ExampleNewContainer() {
	container := di.NewContainer()
	fmt.Println(container)
}
