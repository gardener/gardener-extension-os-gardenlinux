#!/bin/bash

set -e

LC_ALL=C sed -E -n 's/.*k8s\.io\/client-go => k8s\.io\/client-go (.+)/\1/p' go.mod 
#gardener_ver="$(sed -E -n 's/\s*github.com\/gardener\/gardener\s+(\S+)/\1/p' "${REPO_DIR}/go.mod" )"
#echo "$gardener_ver"
#text="$(curl  "https://api.github.com/repos/gardener/gardener/contents/go.mod?ref=v1.60.0" --header "authorization: Bearer ${GH_TOKEN}" )"
#echo "$text"