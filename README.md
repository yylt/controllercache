# controllercache

[![codecov](https://codecov.io/gh/yylt/controllercache/branch/master/graph/badge.svg)](https://codecov.io/gh/yylt/controllercache)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylt/controllercache)](https://goreportcard.com/report/github.com/yylt/controllercache)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylt/controllercache?status.svg)](https://pkg.go.dev/github.com/yylt/controllercache?tab=doc)


controllercache is mainly used as a kubernetes cache and is implemented in the red box below


![cache](doc/client.png)


## install

To install Gin package, you need to install Go and set your Go workspace first.

1. The first need [Go](https://golang.org/) installed (**version 1.16+ is required**)

```sh
$ go get -u github.com/yylt/controllercache
```

1. Import it in your code:

```go
import "github.com/yylt/controllercache"
```

## Quick start

```sh
# assume the following codes in example.go file
$ cat example/get/example.go
```

```go
package main

import (
    "fmt"
    "k8s.io/client-go/tools/clientcmd"
    "github.com/yylt/controllercache"
    corev1 "k8s.io/api/core/v1"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "kubeconfig path")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
    if err != nil {
        panic(err)
    }

    var n = &corev1.Node{}
    reader.CacheForObject(n, "")
    reader.Get(context.TODO(), types.NamespacedName{
        Name: "default",
    }, n)
    fmt.Printf("namespace default: %v\n", n)
}
```
