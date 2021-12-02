package controllercache

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type item struct {
	ctrlcache.Cache
	namespace string
}

type itemer interface {
	Namespace() string
}

func (i *item) Namespace() string {
	return i.namespace
}

// clientcache: with cache Get/List
type ClientCache struct {
	// reuse interface
	ctrlclient.Client

	// config which create informer client
	config *rest.Config

	// context
	ctx    context.Context
	stopfn func()

	scheme *runtime.Scheme
	// schema should add before start
	mu sync.RWMutex

	multis map[schema.GroupVersionKind]ctrlcache.Cache
}

func NewCacheClient(config *rest.Config, scheme *runtime.Scheme) *ClientCache {
	var (
		lisscheme *runtime.Scheme
	)
	if scheme == nil {
		lisscheme = runtime.NewScheme()
		err := clientgoscheme.AddToScheme(lisscheme)
		if err != nil {
			panic(err)
		}
	} else {
		lisscheme = scheme
	}

	cli, err := ctrlclient.New(config, ctrlclient.Options{
		Scheme: lisscheme,
	})
	if err != nil {
		panic(err)
	}
	ctx, cancle := context.WithCancel(context.Background())
	var once sync.Once
	stopfn := func() {
		once.Do(cancle)
	}
	return &ClientCache{
		scheme: lisscheme,
		stopfn: stopfn,
		ctx:    ctx,
		Client: cli,
		config: config,
		multis: make(map[schema.GroupVersionKind]ctrlcache.Cache),
	}
}

// NOT SUPPORT ListObject
// namespace is optional, listen resource on all namespace if namespace is ""
func (c *ClientCache) CacheForObject(obj ctrlclient.Object, namespace string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	gvk, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return err
	}
	if strings.HasSuffix(gvk.Kind, "List") {
		return fmt.Errorf("could not add ListObject")
	}
	option := ctrlcache.Options{
		Scheme: c.scheme,
	}
	if namespace != "" {
		option.Namespace = namespace
	}

	reader, err := ctrlcache.New(c.config, option)
	if err != nil {
		return err
	}
	go func() {
		//TODO record if want to stop watch
		reader.Start(context.TODO()) // nolint
	}()
	if namespace != "" {
		one := &item{
			namespace: namespace,
			Cache:     reader,
		}
		c.multis[gvk] = one
	} else {
		c.multis[gvk] = reader
	}

	return nil
}

func (c *ClientCache) Stop() {
	c.stopfn()
}

func (c *ClientCache) inCache(gvk schema.GroupVersionKind, ns string) ctrlcache.Cache {
	news := gvk.String()

	for k, v := range c.multis {
		if k.String() == news {
			cacheitem, ok := v.(itemer)
			if ns != "" && ok {
				if ns == cacheitem.Namespace() {
					return v
				}
			}
		}
	}
	return nil
}

func (c *ClientCache) Get(ctx context.Context, nsname ctrlclient.ObjectKey, obj ctrlclient.Object) error {
	gvk, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	reader := c.inCache(gvk, nsname.Namespace)
	if reader != nil {
		if !reader.WaitForCacheSync(ctx) {
			klog.Warningf("%s could not sync, but still use cache data", gvk.String())
		}
		return reader.Get(ctx, nsname, obj)
	}
	return c.Client.Get(ctx, nsname, obj)
}

func (c *ClientCache) List(ctx context.Context, obj ctrlclient.ObjectList, opts ...ctrlclient.ListOption) error {
	gvk, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return err
	}
	if strings.HasSuffix(gvk.Kind, "List") {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}
	var option = &ctrlclient.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(option)
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	reader := c.inCache(gvk, option.Namespace)
	if reader != nil {
		if !reader.WaitForCacheSync(ctx) {
			klog.Warningf("%s could not sync, but still use cache data", gvk.String())
		}
		return reader.List(ctx, obj, opts...)
	}
	return c.Client.List(ctx, obj, opts...)
}
