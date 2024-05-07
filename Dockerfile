############# builder
FROM golang:1.22.3 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-os-gardenlinux
COPY . .
RUN make install

############# gardener-extension-os-gardenlinux
FROM gcr.io/distroless/static-debian11:nonroot AS gardener-extension-os-gardenlinux
WORKDIR /

COPY --from=builder /go/bin/gardener-extension-os-gardenlinux /gardener-extension-os-gardenlinux
ENTRYPOINT ["/gardener-extension-os-gardenlinux"]
