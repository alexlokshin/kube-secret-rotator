#!/bin/sh
env GOOS=linux GOARCH=arm go build
docker build -t kube-secret-rotator:latest . 
docker tag kube-secret-rotator:latest alexlokshin/kube-secret-rotator:latest
docker push alexlokshin/kube-secret-rotator:latest
kubectl apply -f rotator-ds.yml