package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type secretDef struct {
	name      string
	namespace string
	key       string
	strategy  string
}

type statusResponse struct {
	Status        string `json:"Status,omitempty"`
	RotationCount int    `json:"RotationCount,omitempty"`
}

var chars = []rune("01234567890$%#!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var rotationCount = 0

func betterPanic(message string) {
	fmt.Printf("%s\n\n", message)
	os.Exit(1)
}

func randomizeString(n int) string {
	byteArray := make([]rune, n)
	for i := range byteArray {
		byteArray[i] = chars[rand.Intn(len(chars))]
	}
	return strings.Replace(b64.StdEncoding.EncodeToString([]byte(string(byteArray))), "=", "", -1)
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	var response = statusResponse{Status: "OK", RotationCount: rotationCount}
	json.NewEncoder(w).Encode(response)
}

func rotate(frequency int, secretDefs []secretDef) {
	config, err := rest.InClusterConfig()
	if err != nil {
		betterPanic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		betterPanic(err.Error())
	}

	for {
		for i := 0; i < len(secretDefs); i++ {
			_, err := clientset.CoreV1().Namespaces().Get(secretDefs[i].namespace, metav1.GetOptions{})
			if err != nil {
				log.Printf("Cannot get the namespace %s, skipping secret creation for now.\n", secretDefs[i].namespace)
				continue
			}

			secret, err := clientset.CoreV1().Secrets(secretDefs[i].namespace).Get(secretDefs[i].name, metav1.GetOptions{})
			log.Printf("Rotating secret %s.%s.\n", secretDefs[i].namespace, secretDefs[i].name)

			newValue := ([]byte)(randomizeString(40))
			t := time.Now()

			if err != nil {
				log.Printf("Error: %s\n", err)
				dataMap := make(map[string][]byte)
				if "retainPrev" == secretDefs[i].strategy {
					dataMap[secretDefs[i].key+"_PREV"] = newValue
				}
				dataMap[secretDefs[i].key] = newValue

				annotations := make(map[string]string)
				annotations["kube-secret-rotator/rotated"] = t.Format(time.RFC850)

				secret = &corev1.Secret{
					Type: corev1.SecretTypeOpaque,
					ObjectMeta: metav1.ObjectMeta{
						Name:        secretDefs[i].name,
						Namespace:   secretDefs[i].namespace,
						Annotations: annotations,
					},
					Data: dataMap,
				}
				log.Printf("Secret %s.%s doesn't exist. Creating.\n", secretDefs[i].namespace, secretDefs[i].name)
				secret, err = clientset.CoreV1().Secrets(secretDefs[i].namespace).Create(secret)
				if err != nil {
					log.Printf("Failed to create secret: %s\n", err.Error())
				}

				rotationCount++
			} else {
				if secret == nil {
					log.Printf("No existing secret found.\n")
					continue
				}
				currentValue, err := b64.StdEncoding.DecodeString(string(secret.Data[secretDefs[i].key]))
				log.Printf("Current value of the secret %s.%s->%s is %s.\n", secretDefs[i].namespace, secretDefs[i].name, secretDefs[i].key, string(currentValue))
				if "retainPrev" == secretDefs[i].strategy {
					secret.Data[secretDefs[i].key+"_PREV"] = secret.Data[secretDefs[i].key]
				}
				secret.Data[secretDefs[i].key] = newValue
				secret.ObjectMeta.Annotations["kube-secret-rotator/rotated"] = t.Format(time.RFC850)
				secret, err = clientset.CoreV1().Secrets(secretDefs[i].namespace).Update(secret)
				if err != nil {
					log.Printf("Failed to update secret: %s\n", err.Error())
				}
				rotationCount++
			}
		}

		time.Sleep(time.Duration(frequency) * time.Minute)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var secretDefs = []secretDef{}
	fmt.Printf("\nKubernetes secret rotator.\n")

	secretArg := flag.String("secret", "", "SECRET_NAME,NAMESPACE,KEY,STRATEGY[|SECRET_NAME,NAMESPACE,KEY,STRATEGY]")
	freqArg := flag.Int("frequency", 60, "Rotation frequency, minutes")
	flag.Parse()

	frequency := *freqArg

	if frequency < 1 {
		betterPanic("Invalid frequency specified.")
	}

	sequences := strings.Split(*secretArg, "|")
	if len(sequences) < 1 {
		betterPanic("At least one secret sequence has to be specified.")
	}

	for i := 0; i < len(sequences); i++ {
		parts := strings.Split(sequences[i], ",")
		if len(parts) != 4 {
			betterPanic("Invalid specification for the secret. Valid sequence is SECRET_NAME,NAMESPACE,KEY,STRATEGY. For example: tempsecret,default,somekey,retainPrev")
		}
		secret := secretDef{name: parts[0], namespace: parts[1], key: parts[2], strategy: parts[3]}
		secretDefs = append(secretDefs, secret)

		log.Printf("Rotating secret `%s` in the namespace of `%s` every %d minutes.\n", secret.name, secret.namespace, frequency)
	}

	// Kicks off the endless loop
	go rotate(frequency, secretDefs)

	log.Println("Starting a web server on 0.0.0.0:8080")
	router := mux.NewRouter()
	router.HandleFunc("/", getStatus).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
