/*
Package controllercache implements
	* cache for kubernetes resource, use container-runtime package.
	* iter resource.


Example:
	import (
		"k8s.io/client-go/tools/clientcmd"
		"github.com/yylt/controllercache"
		corev1 "k8s.io/api/core/v1"
	)

	func main() {
		kubeconfig := ""
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}

		var n = &corev1.Node{}
		reader.CacheForObject(n, "")
		reader.Get(context.TODO(), types.NamespacedName{
			Name: "default",
		}, n)
		fmt.Printf("default ns: %#v\n", n)
	}

*/
package controllercache
