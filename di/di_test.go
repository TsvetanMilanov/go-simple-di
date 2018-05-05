package di

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type rootDependency struct {
	First     *firstLevelDependency `di:""`
	Pointer   *pointerDependency    `di:""`
	Interface worker                `di:""`
}

type firstLevelDependency struct {
	Second             *secondLevelDependency `di:""`
	PointerSecondLevel *pointerDependency     `di:""`
}

type secondLevelDependency struct {
	PointerThirdLevel   *pointerDependency `di:""`
	InterfaceThirdLevel worker             `di:""`
}

type pointerDependency struct {
	value int
}

type worker interface {
	Work() string
}

type builder struct {
	work string
}

type named struct {
	Struct    *pointerDependency `di:"name=test1"`
	Interface worker             `di:"name=test2"`
}

type first struct {
	S *second            `di:""`
	P *pointerDependency `di:""`
}

type second struct {
	F *first `di:""`
}

type third struct {
	S *second `di:""`
}

func (b *builder) Work() string { return b.work }

func TestDependencyInjection(t *testing.T) {
	Convey("Container", t, func() {
		Convey("Resolve", func() {
			Convey("Should resolve", func() {
				Convey("dependencies recursively.", func() {
					c := NewContainer()
					value := 5
					work := "Work"
					root := &Dependency{Value: new(rootDependency)}
					first := &Dependency{Value: new(firstLevelDependency)}
					second := &Dependency{Value: new(secondLevelDependency)}
					ptr := &Dependency{Value: &pointerDependency{value: value}}
					b := &Dependency{Value: &builder{work: work}}

					err := c.Register(root, first, second, ptr, b)
					So(err, ShouldBeNil)

					res := new(rootDependency)
					err = c.Resolve(res)

					So(err, ShouldBeNil)

					// Struct assertions.
					So(res.Pointer, ShouldNotBeNil)
					So(res.Pointer.value, ShouldEqual, value)
					So(res.First, ShouldNotBeNil)
					So(res.First.PointerSecondLevel, ShouldNotBeNil)
					So(res.First.PointerSecondLevel.value, ShouldEqual, value)
					So(res.First.Second, ShouldNotBeNil)
					So(res.First.Second.PointerThirdLevel, ShouldNotBeNil)
					So(res.First.Second.PointerThirdLevel.value, ShouldEqual, value)

					// Interface assertions.
					So(res.Interface, ShouldNotBeNil)
					So(res.Interface.Work(), ShouldEqual, work)
					So(res.First.Second.InterfaceThirdLevel, ShouldNotBeNil)
					So(res.First.Second.InterfaceThirdLevel.Work(), ShouldEqual, work)
				})
				Convey("interface values.", func() {
					c := NewContainer()
					w := "Builder"
					err := c.Register(&Dependency{Value: &builder{work: w}})
					So(err, ShouldBeNil)

					res := new(worker)
					err = c.Resolve(res)
					So(err, ShouldBeNil)
					So((*res).Work(), ShouldEqual, w)
				})
				Convey("named structs and interfaces.", func() {
					c := NewContainer()
					w := "Builder"
					v := 50

					n := &Dependency{Value: new(named)}
					p := &Dependency{Value: &pointerDependency{value: v}, Name: "test1"}
					b := &Dependency{Value: &builder{work: w}, Name: "test2"}
					err := c.Register(n, p, b)
					So(err, ShouldBeNil)

					res := new(named)
					err = c.Resolve(res)
					So(err, ShouldBeNil)
					So(res.Struct, ShouldNotBeNil)
					So(res.Interface, ShouldNotBeNil)
					So(res.Struct.value, ShouldEqual, v)
					So(res.Interface.Work(), ShouldEqual, w)
				})
				Convey("self if it's registered to its dependencies", func() {
					Convey("direct resolve", func() {
						c := NewContainer()

						err := c.Register(&Dependency{Value: c})
						So(err, ShouldBeNil)

						self := new(Container)
						err = c.Resolve(self)

						So(err, ShouldBeNil)
						So(c, ShouldResemble, self)
						So(*c, ShouldResemble, *self)
					})
					Convey("resolve in struct property.", func() {
						c := NewContainer()
						type selfResolvable struct {
							C *Container `di:""`
						}

						err := c.Register(
							&Dependency{Value: c},
							&Dependency{Value: new(selfResolvable)},
						)
						So(err, ShouldBeNil)

						self := new(selfResolvable)
						err = c.Resolve(self)

						So(err, ShouldBeNil)
						So(c, ShouldResemble, self.C)
						So(*c, ShouldResemble, *self.C)
					})
				})
				Convey("circular dependencies.", func() {
					c := NewContainer()
					v := 100
					err := c.Register(
						&Dependency{Value: new(first)},
						&Dependency{Value: new(second)},
						&Dependency{Value: new(third)},
						&Dependency{Value: &pointerDependency{value: v}},
					)
					So(err, ShouldBeNil)

					t := new(third)
					err = c.Resolve(t)
					So(err, ShouldBeNil)
					So(t.S, ShouldNotBeNil)

					f := new(first)
					err = c.Resolve(f)
					So(err, ShouldBeNil)
					So(f.S, ShouldNotBeNil)
					So(f.P, ShouldNotBeNil)
					So(f.P.value, ShouldEqual, v)

					s := new(second)
					err = c.Resolve(s)
					So(err, ShouldBeNil)
					So(s.F, ShouldNotBeNil)
				})
			})
			Convey("Should fail to resolve", func() {
				Convey("unnamed dependencies for named structs.", func() {
					c := NewContainer()
					w := "Builder"
					v := 50

					n := &Dependency{Value: new(named)}
					p := &Dependency{Value: &pointerDependency{value: v}}
					b := &Dependency{Value: &builder{work: w}, Name: "test2"}
					err := c.Register(n, p, b)
					So(err, ShouldBeNil)

					res := new(named)
					err = c.Resolve(res)
					So(err, ShouldBeError, "[*di.named] unable to find registered dependency: Struct")
				})
				Convey("unnamed dependencies for named interfaces.", func() {
					c := NewContainer()
					w := "Builder"
					v := 50

					n := &Dependency{Value: new(named)}
					p := &Dependency{Value: &pointerDependency{value: v}, Name: "test1"}
					b := &Dependency{Value: &builder{work: w}}
					err := c.Register(n, p, b)
					So(err, ShouldBeNil)

					res := new(named)
					err = c.Resolve(res)
					So(err, ShouldBeError, "[*di.named] unable to find registered dependency: Interface")
				})
				Convey("when not all dependencies are registered.", func() {
					c := NewContainer()
					root := &Dependency{Value: new(rootDependency)}
					first := &Dependency{Value: new(firstLevelDependency)}
					second := &Dependency{Value: new(secondLevelDependency)}
					ptr := &Dependency{Value: new(pointerDependency)}

					err := c.Register(root, first, second, ptr)
					So(err, ShouldBeNil)

					res := new(rootDependency)
					err = c.Resolve(res)

					So(err, ShouldBeError, "[*di.rootDependency] [*di.firstLevelDependency] [*di.secondLevelDependency] unable to find registered dependency: InterfaceThirdLevel")
				})
				Convey("unexported properties.", func() {
					type unexp struct {
						iAmNotExported *pointerDependency `di:""`
					}

					c := NewContainer()
					err := c.Register(
						&Dependency{Value: new(unexp)},
						&Dependency{Value: new(pointerDependency)},
					)
					So(err, ShouldBeNil)

					r := new(unexp)
					err = c.Resolve(r)

					So(err, ShouldBeError, "[*di.unexp] cannot set field iAmNotExported")
					So(r.iAmNotExported, ShouldBeNil)
				})
				Convey("when the tag is invalid.", func() {
					c := NewContainer()
					type invalidTag struct {
						P *pointerDependency `di:"name="`
					}
					err := c.Register(
						&Dependency{Value: new(invalidTag)},
						&Dependency{Value: new(pointerDependency)},
					)
					So(err, ShouldBeNil)

					r := new(invalidTag)
					err = c.Resolve(r)

					So(err, ShouldBeError, "[*di.invalidTag] "+getInvalidTagErr("name=").Error())
				})
			})
			Convey("Should NOT resolve fields without tags", func() {
				c := NewContainer()
				type notag struct {
					ResolveMe     *pointerDependency `di:""`
					DontResolveMe *pointerDependency
					OtherTags     *pointerDependency `json:"otherTags"`
				}

				v := 75
				err := c.Register(
					&Dependency{Value: new(notag)},
					&Dependency{Value: &pointerDependency{value: v}},
				)
				So(err, ShouldBeNil)

				res := new(notag)
				err = c.Resolve(res)

				So(err, ShouldBeNil)
				So(res.DontResolveMe, ShouldBeNil)
				So(res.OtherTags, ShouldBeNil)
				So(res.ResolveMe, ShouldNotBeNil)
				So(res.ResolveMe.value, ShouldEqual, v)
			})
			Convey("Should validate the out parameter to be pointer for", func() {
				Convey("structs.", func() {
					c := NewContainer()
					err := c.Resolve(pointerDependency{})
					So(err, ShouldBeError, "the out parameter must be a pointer")
				})
				Convey("interfaces.", func() {
					c := NewContainer()
					var w worker
					err := c.Resolve(w)
					So(err, ShouldBeError, "the out parameter must be a pointer")
				})
			})
		})

		Convey("ResolveByName", func() {
			Convey("Should resolve", func() {
				Convey("structs.", func() {
					c := NewContainer()
					w := "Builder"
					err := c.Register(&Dependency{Value: &builder{work: w}, Name: "test"})
					So(err, ShouldBeNil)

					res := new(builder)
					err = c.ResolveByName("test", res)
					So(err, ShouldBeNil)
					So(res.Work(), ShouldEqual, w)
				})
				Convey("interfaces.", func() {
					c := NewContainer()
					w := "Builder"
					err := c.Register(&Dependency{Value: &builder{work: w}, Name: "test"})
					So(err, ShouldBeNil)

					res := new(worker)
					err = c.ResolveByName("test", res)
					So(err, ShouldBeNil)
					So((*res).Work(), ShouldEqual, w)
				})
			})
			Convey("Should NOT resolve", func() {
				Convey("unnamed structs.", func() {
					c := NewContainer()
					w := "Builder"
					err := c.Register(&Dependency{Value: &builder{work: w}})
					So(err, ShouldBeNil)

					res := new(builder)
					err = c.ResolveByName("test", res)
					So(err, ShouldBeError, "unable to find registered dependency: *di.builder")
				})
				Convey("unnamed structs which implement interfaces.", func() {
					c := NewContainer()
					w := "Builder"
					err := c.Register(&Dependency{Value: &builder{work: w}})
					So(err, ShouldBeNil)

					res := new(worker)
					err = c.ResolveByName("test", res)
					So(err, ShouldBeError, "unable to find registered dependency: *di.worker")
				})
			})
		})

		Convey("Register", func() {
			Convey("Should validate the dependency value to be pointer.", func() {
				c := NewContainer()
				err := c.Register(&Dependency{Value: pointerDependency{}})

				So(err, ShouldBeError, "di.pointerDependency should be pointer or interface")
			})
			Convey("Should check for duplicate dependency registration,", func() {
				c := NewContainer()
				d := &Dependency{Value: new(pointerDependency)}
				err := c.Register(d, d)

				So(err, ShouldBeError, "duplicate dependency: -*di.pointerDependency-ptr")
			})
		})

		Convey("ResolveAll", func() {
			Convey("Should resolve all distinct hierarchies.", func() {
				c := NewContainer()
				v := 100
				w := "ResolveAll"
				r1Value := new(struct {
					P *pointerDependency `di:""`
				})
				r2Value := new(struct {
					I worker `di:""`
				})
				r1 := &Dependency{Value: r1Value}
				r2 := &Dependency{Value: r2Value}
				p := &Dependency{Value: &pointerDependency{value: v}}
				b := &Dependency{Value: &builder{work: w}}
				err := c.Register(r1, r2, p, b)

				So(err, ShouldBeNil)

				err = c.ResolveAll()

				So(err, ShouldBeNil)
				So(r1Value.P, ShouldNotBeNil)
				So(r1Value.P.value, ShouldEqual, v)
				So(r2Value.I, ShouldNotBeNil)
				So(r2Value.I.Work(), ShouldEqual, w)
			})
			Convey("Should return resolve error.", func() {
				c := NewContainer()
				d := &Dependency{Value: new(firstLevelDependency)}
				c.Register(d)

				err := c.ResolveAll()

				So(err, ShouldBeError, "[*di.firstLevelDependency] unable to find registered dependency: Second")
			})
		})
	})
}
