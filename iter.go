package controllercache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	IterFactoryPool = sync.Pool{
		New: func() interface{} {
			return newIterFactory()
		},
	}
)

func GetIter() *IterFactory {
	factory := IterFactoryPool.New().(*IterFactory)
	factory.Reset()
	return factory
}

func PutIter(i *IterFactory) {
	IterFactoryPool.Put(i)
}

type IterFactory struct {
	options []ctrlclient.ListOption

	object ctrlclient.ObjectList
	ctx    context.Context

	processFn func(runtime.Object) error
}

func newIterFactory() *IterFactory {
	return &IterFactory{
		ctx: context.Background(),
	}
}

func (i *IterFactory) Reset() *IterFactory {
	i.options = i.options[:0]
	i.processFn = nil
	i.object = nil
	return i
}

func (i *IterFactory) Namespace(ns string) *IterFactory {
	i.options = append(i.options, ctrlclient.InNamespace(ns))
	return i
}

func (i *IterFactory) Labels(labels map[string]string) *IterFactory {
	if labels != nil {
		i.options = append(i.options, ctrlclient.MatchingLabels(labels))
	}
	return i
}

func (i *IterFactory) Fiedls(fields map[string]string) *IterFactory {
	if fields != nil {
		i.options = append(i.options, ctrlclient.MatchingFields(fields))
	}
	return i
}

func (i *IterFactory) Object(object ctrlclient.ObjectList) *IterFactory {
	i.object = object
	return i
}

func (i *IterFactory) Fn(fn func(runtime.Object) error) *IterFactory {
	i.processFn = fn
	return i
}

func (i *IterFactory) preDoCheck() error {
	if i.object == nil {
		return fmt.Errorf("iter factory not set ObjectList")
	}
	if i.processFn == nil {
		return fmt.Errorf("iter factory not set function")
	}
	return nil
}

func (i *IterFactory) Do(cli ctrlclient.Client, timeout time.Duration) error {
	var (
		err error
	)
	err = i.preDoCheck()
	if err != nil {
		return err
	}
	ctx, cancle := context.WithTimeout(i.ctx, timeout)
	defer cancle()

	err = cli.List(ctx, i.object, i.options...)
	if err != nil {
		return err
	}
	return i.processFn(i.object)
}
