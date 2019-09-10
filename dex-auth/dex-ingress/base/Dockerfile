FROM alpine

MAINTAINER Krishna Durai <krishadurai20@gmail.com>

ENV KUBE_LATEST_VERSION="v1.15.0"

RUN apk add --update ca-certificates \
 && apk add --update -t deps curl \
 && apk add --no-cache bash \
 && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl \
 && apk add --no-cache openssl \
 && apk del --purge deps \
 && rm /var/cache/apk/* \
 && mkdir /var/ingress-configs/

ENTRYPOINT ["bash"]
CMD ["--help"]
