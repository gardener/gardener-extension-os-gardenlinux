#!/bin/bash

go mod edit -require "github.com/gardener/gardener@v1.6.0"

gardener_ver="$(sed -E -n 's/\s*github.com\/gardener\/gardener\s+(\S+)/\1/p' go.mod)"
gardener_replace="$(curl  "https://api.github.com/repos/gardener/gardener/contents/go.mod?ref=${gardener_ver}" | jq -r ".content" | base64 --decode | sed -E -n 's/.*k8s\.io\/client-go => k8s\.io\/client-go (.+)/\1/p')"
go mod edit -replace "k8s.io/client-go=k8s.io/client-go@${gardener_replace}"

go mod edit -require "github.com/gardener/gardener@v1.60.0"