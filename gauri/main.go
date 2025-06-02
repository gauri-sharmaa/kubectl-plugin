package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	if len(os.Args) < 3 || !strings.HasPrefix(os.Args[1], "get") {
		fmt.Println("Usage: kubectl multi get <resource> [--namespace <ns>]")
		os.Exit(1)
	}

	resource := os.Args[2]
	namespace := ""

	fs := flag.NewFlagSet("multi", flag.ExitOnError)
	fs.StringVar(&namespace, "namespace", "", "Target namespace (optional)")
	_ = fs.Parse(os.Args[3:])

	fmt.Println("\nLocal Cluster Output:")
	localArgs := []string{"get", resource}
	if namespace != "" {
		localArgs = append(localArgs, "-n", namespace)
	}
	cmd := exec.Command("kubectl", localArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	configAccess := clientcmd.NewDefaultPathOptions()
	rawConfig, err := configAccess.GetStartingConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load kubeconfig: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nDiscovered Contexts:")
	for name := range rawConfig.Contexts {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println("\nRemote Cluster Results:")
	for name := range rawConfig.Contexts {
		fmt.Printf("Cluster: %s\n", name)

		config, err := clientcmd.NewNonInteractiveClientConfig(*rawConfig, name, &clientcmd.ConfigOverrides{}, configAccess).ClientConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Could not load kubeconfig for context %s: %v\n", name, err)
			continue
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Could not connect to context %s: %v\n", name, err)
			continue
		}

		if isClusterScoped(resource) {
			dynClient, err := dynamic.NewForConfig(config)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Failed to create dynamic client for context %s: %v\n", name, err)
				continue
			}
			gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: resource}
			items, err := dynClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Failed to get %s in context %s: %v\n", resource, name, err)
				continue
			}
			for _, obj := range items.Items {
				fmt.Printf("    - %s\n", obj.GetName())
			}
		} else {
			ns := namespace
			if ns == "" {
				ns = "default"
			}
			raw, err := clientset.CoreV1().RESTClient().
				Get().
				Namespace(ns).
				Resource(resource).
				DoRaw(context.TODO())
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Failed to get %s in ns %s for context %s: %v\n", resource, ns, name, err)
				continue
			}
			prettyPrintJSON(raw)
		}
	}
}

func prettyPrintJSON(raw []byte) {
	var obj interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		fmt.Println(string(raw))
		return
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Println(string(raw))
		return
	}
	fmt.Println(string(pretty))
}

func isClusterScoped(res string) bool {
	clusterScoped := map[string]bool{
		"nodes":             true,
		"namespaces":        true,
		"persistentvolumes": true,
	}
	return clusterScoped[res]
}
