#!/bin/bash
set -e

gardener_ver="$(go list -m -f '{{.Version}}' github.com/gardener/gardener)"
gardener_go_mod="$(curl  "https://api.github.com/repos/gardener/gardener/contents/go.mod?ref=${gardener_ver}" --header "authorization: Bearer ${GH_TOKEN}" | jq -r ".content" | base64 --decode )"
client_go_replace="$(echo "$gardener_go_mod"| sed -E -n 's/.*k8s\.io\/client-go => k8s\.io\/client-go (.+)/\1/p')"
go mod edit -replace "k8s.io/client-go=k8s.io/client-go@${client_go_replace}"