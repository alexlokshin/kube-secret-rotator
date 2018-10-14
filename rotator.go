package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type secretDef struct {
	name      string
	namespace string
	key       string
}

func main() {
	var secretDefs = []secretDef{}
	fmt.Printf("Kubernetes secret rotator.\n")

	secretArg := flag.String("secret", "", "SECRET_NAME,NAMESPACE,KEY[|SECRET_NAME,NAMESPACE,KEY]")
	freqArg := flag.Int("frequency", 60, "Rotation frequency, minutes")
	flag.Parse()

	frequency := *freqArg

	sequences := strings.Split(*secretArg, "|")
	for i := 0; i < len(sequences); i++ {
		parts := strings.Split(sequences[i], ",")
		if len(parts) != 4 {
			panic("Invalid specification for the secret. Valid sequence is SECRET_NAME,NAMESPACE,FREQ.")

		}
		secret := secretDef{name: parts[0], namespace: parts[1], key: parts[2]}
		secretDefs = append(secretDefs, secret)
		fmt.Printf("Rotating secret `%s` in the namespace of `%s` every %d minutes.\n", secret.name, secret.namespace, frequency)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		for i := 0; i < len(secretDefs); i++ {
			secret, err := clientset.CoreV1().Secrets(secretDefs[i].namespace).Get(secretDefs[i].name, metav1.GetOptions{})
			newValue, err := uuid.NewV4()
			if err != nil {
				fmt.Printf("%s", err)
				dataMap := make(map[string]string)
				dataMap[secretDefs[i].key] = newValue.String()
				secret = &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretDefs[i].name,
						Namespace: secretDefs[i].namespace,
					},
					StringData: dataMap,
				}
				fmt.Printf("Secret %s.%s doesn't exist.\n", secretDefs[i].namespace, secretDefs[i].name)
				clientset.CoreV1().Secrets(secretDefs[i].namespace).Create(secret)
			} else {
				fmt.Printf("Current value of the secret %s.%s->%s is %s.\n", secretDefs[i].namespace, secretDefs[i].name, secretDefs[i].key, secret.StringData[secretDefs[i].key])
				secret.StringData[secretDefs[i].key] = newValue.String()
				clientset.CoreV1().Secrets(secretDefs[i].namespace).Update(secret)
			}
		}

		time.Sleep(time.Duration(frequency) * time.Minute)
	}
}
