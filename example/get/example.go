package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/yylt/controllercache"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
)

func noderead(reader *controllercache.ClientCache) {
	var n = &corev1.Node{}
	reader.CacheForObject(n, "")
	reader.Get(context.TODO(), types.NamespacedName{
		Name: "default",
	}, n)
	fmt.Printf("namespace default: %v\n", n)
}

func deployread(reader *controllercache.ClientCache) {
	var (
		dep  = &appsv1.Deployment{}
		ns   = "kube-system"
		name = "coredns"
	)
	reader.CacheForObject(dep, ns)
	reader.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: ns,
	}, dep)
	fmt.Printf("deploy %s/%s: %v\n", ns, name, dep)
}

func main() {
	kubeconfig := flag.String("kubeconfig", "", "kubeconfig file path")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	reader := controllercache.NewCacheClient(config, nil)
	noderead(reader)
	deployread(reader)
}
