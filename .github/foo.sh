#!/bin/bash

sed -E -n 's/.*k8s\.io\/client-go => k8s\.io\/client-go (.+)/\1/p' go.mod 
echo "sep"

sed -E -n 's/.*github\.com\/gardener\/gardener (\S+)/\1/p' go.mod
echo "sep"

sed -E -n 's/.*github\.com\/gardener\/gardener (.+)/\1/p' go.mod
#gardener_ver="$(sed -E -n 's/\s*github.com\/gardener\/gardener\s+(\S+)/\1/p' "${REPO_DIR}/go.mod" )"
#echo "$gardener_ver"
#text="$(curl  "https://api.github.com/repos/gardener/gardener/contents/go.mod?ref=v1.60.0" --header "authorization: Bearer ${GH_TOKEN}" | jq -r ".content" | base64 --decode )"
#echo "$text"