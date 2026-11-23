############# builder
FROM golang:1.27.4@sha256:205c7636d74e0ffbe3d007162df7d74f859f5a8c8322dfb9f907977686c126cf AS builder

ARG EFFECTIVE_VERSION
WORKDIR /go/src/github.com/gardener/gardener-extension-runtime-gvisor
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# binaries-installer
FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf AS binaries-installer

COPY hack/ hack/
COPY GVISOR_VERSION ./
RUN hack/install-binaries.sh $(cat GVISOR_VERSION)

############# gardener-extension-runtime-gvisor
FROM gcr.io/distroless/static-debian13:nonroot@sha256:8ce6dd87e9823c5b48dcfb2c8604714ccb351ce59607327bcc4957448da279b5 AS gardener-extension-runtime-gvisor
WORKDIR /

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-runtime-gvisor /gardener-extension-runtime-gvisor
ENTRYPOINT ["/gardener-extension-runtime-gvisor"]

############# gardener-extension-runtime-gvisor-installation for the installation daemonSet
FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf AS gardener-extension-runtime-gvisor-installation

COPY --from=binaries-installer /usr/local/bin/containerd-shim-runsc-v1 /var/content/containerd-shim-runsc-v1
COPY --from=binaries-installer /usr/local/bin/runsc /var/content/runsc
COPY --from=binaries-installer /usr/local/bin/gvisor-bin /var/content/gvisor-bin
