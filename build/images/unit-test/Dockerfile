# This dockerfile is used to build the unit-test image.
# It is used to run the unit-test container.

FROM golang:1.17

RUN apt update && \
    apt install -y --no-install-recommends openvswitch-switch && \
    apt clean && rm -rf /var/lib/apt/lists/* /var/cache/apt/*

RUN mkdir -p /usr/local/kubebuilder/bin/

ENV KUBERNETES_RELEASE_URL=https://storage.googleapis.com/kubernetes-release/release
ENV KUBERNETES_VERSION=v1.18.17
RUN curl -L ${KUBERNETES_RELEASE_URL}/${KUBERNETES_VERSION}/bin/linux/amd64/kube-apiserver -o /usr/local/kubebuilder/bin/kube-apiserver

ENV ETCD_RELEASE_URL=https://github.com/etcd-io/etcd/releases/download
ENV ETCD_VERSION=v3.5.0
RUN curl -L ${ETCD_RELEASE_URL}/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VERSION}.tar && \
  tar xf /tmp/etcd-${ETCD_VERSION}.tar -C /usr/local/kubebuilder/bin --strip-components=1 --extract etcd-${ETCD_VERSION}-linux-amd64/etcd && \
  rm -rf /tmp/etcd-${ETCD_VERSION}.tar
