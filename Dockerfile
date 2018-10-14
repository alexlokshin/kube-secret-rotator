FROM alpine:3.8
ADD ./kube-secret-rotator .
ENTRYPOINT [ "./kube-secret-rotator" ]