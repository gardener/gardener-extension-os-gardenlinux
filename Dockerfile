############# builder
FROM golang:1.27.4@sha256:87f49b77a7b6a99ad7dc656be010ae453322459dd2ed902ac177649a302bc101 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-os-gardenlinux
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make install

############# gardener-extension-os-gardenlinux
FROM gcr.io/distroless/static-debian13:nonroot@sha256:8ce6dd87e9823c5b48dcfb2c8604714ccb351ce59607327bcc4957448da279b5 AS gardener-extension-os-gardenlinux
WORKDIR /

COPY --from=builder /go/bin/gardener-extension-os-gardenlinux /gardener-extension-os-gardenlinux
ENTRYPOINT ["/gardener-extension-os-gardenlinux"]
