#
# Kube Secret Rotator
#
# An alpine 3.8 based image with a kube-secret-rotator in it.
#
# https://github.com/alexlokshin/kube-secret-rotator
#

FROM alpine:3.8
ADD ./kube-secret-rotator .
ENTRYPOINT [ "/kube-secret-rotator" ]