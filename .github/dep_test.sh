#!/bin/bash

set -e

go mod edit -require "github.com/gardener/gardener@v1.6.0"
#gardener_ver="$(go list -m -f '{{.Version}}' 'github.com/gardener/gardener')"
gardener_ver="$(LC_ALL=C sed -E -n 's/\s*github.com\/gardener\/gardener\s+(\S+)/\1/p' go.mod)"
text="$(curl  "https://api.github.com/repos/gardener/gardener/contents/go.mod?ref=${gardener_ver}" | jq -r ".content" | base64 --decode )"
gardener_replace="$(echo "$text"| LC_ALL=C sed -E -n 's/.*k8s\.io\/client-go => k8s\.io\/client-go (.+)/\1/p')"
go mod edit -replace "k8s.io/client-go=k8s.io/client-go@${gardener_replace}"

go mod edit -require "github.com/gardener/gardener@v1.60.0"