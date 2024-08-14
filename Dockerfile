############# builder
FROM golang:1.23.0 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-os-gardenlinux
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make install

############# gardener-extension-os-gardenlinux
FROM gcr.io/distroless/static-debian11:nonroot AS gardener-extension-os-gardenlinux
WORKDIR /

COPY --from=builder /go/bin/gardener-extension-os-gardenlinux /gardener-extension-os-gardenlinux
ENTRYPOINT ["/gardener-extension-os-gardenlinux"]
