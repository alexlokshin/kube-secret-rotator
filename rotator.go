package main

import (
	b64 "encoding/base64"
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

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

var chars = []rune("01234567890$%#!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomizeString(n int) string {
	byteArray := make([]rune, n)
	for i := range byteArray {
		byteArray[i] = chars[rand.Intn(len(chars))]
	}
	return strings.Replace(b64.StdEncoding.EncodeToString([]byte(string(byteArray))), "=", "", -1)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var secretDefs = []secretDef{}
	fmt.Printf("Kubernetes secret rotator.\n")

	secretArg := flag.String("secret", "", "SECRET_NAME,NAMESPACE,KEY[|SECRET_NAME,NAMESPACE,KEY]")
	freqArg := flag.Int("frequency", 60, "Rotation frequency, minutes")
	flag.Parse()

	frequency := *freqArg

	if frequency < 1 {
		panic("Invalid frequency specified.")
	}

	sequences := strings.Split(*secretArg, "|")
	if len(sequences) < 1 {
		panic("At least one secret sequence has to be specified.")
	}

	for i := 0; i < len(sequences); i++ {
		parts := strings.Split(sequences[i], ",")
		if len(parts) != 3 {
			panic("Invalid specification for the secret. Valid sequence is SECRET_NAME,NAMESPACE,KEY.")
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
			fmt.Printf("Rotating secret %s.%s.\n", secretDefs[i].namespace, secretDefs[i].name)

			newValue := randomizeString(40)
			if err != nil {
				fmt.Printf("%s", err)
				dataMap := make(map[string]string)
				dataMap[secretDefs[i].key+"_PREV"] = newValue
				dataMap[secretDefs[i].key] = newValue

				secret = &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretDefs[i].name,
						Namespace: secretDefs[i].namespace,
					},
					StringData: dataMap,
				}
				fmt.Printf("Secret %s.%s doesn't exist. Creating.\n", secretDefs[i].namespace, secretDefs[i].name)
				secret, err = clientset.CoreV1().Secrets(secretDefs[i].namespace).Create(secret)
				if err != nil {
					fmt.Printf("Failed to create secret: %s\n", err.Error())
				}
			} else {
				fmt.Printf("Current value of the secret %s.%s->%s is %s.\n", secretDefs[i].namespace, secretDefs[i].name, secretDefs[i].key, secret.StringData[secretDefs[i].key])
				secret.StringData[secretDefs[i].key+"_PREV"] = secret.StringData[secretDefs[i].key]
				secret.StringData[secretDefs[i].key] = newValue
				secret, err = clientset.CoreV1().Secrets(secretDefs[i].namespace).Update(secret)
				if err != nil {
					fmt.Printf("Failed to update secret: %s\n", err.Error())
				}
			}
		}

		time.Sleep(time.Duration(frequency) * time.Minute)
	}
}
