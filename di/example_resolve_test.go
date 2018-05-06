package di_test

import (
	"fmt"

	"github.com/TsvetanMilanov/go-simple-di/di"
)

type magic interface {
	Magic() string
}

type magician struct {
	Spell *spell `di:""`
}

func (m *magician) Magic() string {
	return m.Spell.name
}

type spell struct {
	name string
}

func ExampleContainer_Resolve_struct() {
	// create the di container and add all dependencies to it.
	c := di.NewContainer()
	c.Register(
		&di.Dependency{Value: new(magician)},
		&di.Dependency{Value: &spell{name: "fireblast"}},
	)

	// resolve the struct dependency.
	res := new(magician)
	err := c.Resolve(res)
	if err != nil {
		panic(err)
	}

	fmt.Println("Magic:", res.Magic())
	// Output:
	// Magic: fireblast
}
