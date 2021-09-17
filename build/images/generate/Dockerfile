FROM golang:1.17

RUN go install github.com/99designs/gqlgen@v0.13.0 && \
    go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2 && \
    go clean -cache

ENV APISERVER_BOOT_RELEASE_URL=https://github.com/kubernetes-sigs/apiserver-builder-alpha/releases/download/v1.18.0/apiserver-builder-alpha-v1.18.0-linux-amd64.tar.gz
ENV GOROOT=/usr/local/go
RUN wget -qO- $APISERVER_BOOT_RELEASE_URL | tar -C /go -zxf -
