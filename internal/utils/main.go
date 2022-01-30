package utils

import (
	"fmt"
	"os"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Contains represents func for checking if a string is in a list of strings
func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// WriteFile represents func for writing content to a local file
func WriteFile(name, path string, content []byte) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0750); err != nil {
			return err
		}
	}

	f, err := os.Create(path + "/" + name)

	if err != nil {
		return err
	}

	l, err := f.Write(content)

	if err != nil {
		return err
	}

	fmt.Println(l, "bytes written successfully")
	return nil
}

func getClusterConfig() *rest.Config {

	restConfig, err := rest.InClusterConfig()

	if err != nil {

		configLoadRules := clientcmd.NewDefaultClientConfigLoadingRules()
		config, _ := configLoadRules.Load()
		copy := *config

		clientConfig := clientcmd.NewDefaultClientConfig(copy, &clientcmd.ConfigOverrides{})
		restConfig, err = clientConfig.ClientConfig()

		if err != nil {
			panic(err.Error())
		}
	}

	return restConfig
}

func GetTypedKubernetesClient() *kubernetes.Clientset {

	config := getClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}

func GetDynamicKubernetesClient() dynamic.Interface {

	config := getClusterConfig()
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return client
}

func GetDiscoveryKubernetesClient() *discovery.DiscoveryClient {

	config := getClusterConfig()
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err)
	}

	return client
}
