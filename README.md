# go-simple-di
[![GoDoc](https://godoc.org/github.com/TsvetanMilanov/go-simple-di/di?status.svg)](https://godoc.org/github.com/TsvetanMilanov/go-simple-di/di)
[![code-coverage](https://gocover.io/_badge/github.com/TsvetanMilanov/go-simple-di/di)](https://gocover.io/github.com/TsvetanMilanov/go-simple-di/di)
[![Go Report Card](https://goreportcard.com/badge/github.com/TsvetanMilanov/go-simple-di)](https://goreportcard.com/report/github.com/TsvetanMilanov/go-simple-di)
![Go](https://github.com/TsvetanMilanov/go-simple-di/workflows/Go/badge.svg?branch=master)
![Create Release](https://github.com/TsvetanMilanov/go-simple-di/workflows/Create%20Release/badge.svg)

## Contents
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Documentation](#documentation)

## Installation
```shell
go get github.com/TsvetanMilanov/go-simple-di/di
```

## Quick Start
```Go
package main

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

func main() {
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
```

## Documentation
[Godoc](https://godoc.org/github.com/TsvetanMilanov/go-simple-di/di)
