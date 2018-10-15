#!/bin/sh

#docker build -t kube-secret-rotator:latest . 
#docker tag kube-secret-rotator:latest alexlokshin/kube-secret-rotator:latest
#docker push alexlokshin/kube-secret-rotator:latest
kubectl create -f k8s/rbac.yml
kubectl create -f k8s/rotator-ds.yml