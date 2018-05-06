package di_test

import (
	"fmt"

	"github.com/TsvetanMilanov/go-simple-di/di"
)

type breaker interface {
	Break() string
}

type programmer struct {
	Item *breakable `di:""`
}

func (p *programmer) Break() string {
	return p.Item.name
}

type breakable struct {
	name string
}

func ExampleContainer_Resolve_interface() {
	// Create the di container and add all dependencies to it.
	c := di.NewContainer()
	c.Register(
		&di.Dependency{Value: new(programmer)},          // Register the interface implementation.
		&di.Dependency{Value: &breakable{name: "code"}}, // Register other dependencies.
	)

	// Resolve the interface dependency.
	// Due to a limitation in the current go version (go1.9.4), if the out
	// parameter is interface, it should be pointer to interface.
	res := new(breaker)
	err := c.Resolve(res)
	if err != nil {
		panic(err)
	}

	fmt.Println("Breaking:", (*res).Break()) // Deref the interface and use its value.
	// Output:
	// Breaking: code
}
