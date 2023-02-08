#!/bin/bash

gardener_ver="$(sed -E -n 's/\s*github.com\/gardener\/gardener\s+(\S+)/\1/p' go.mod)"
echo "$gardener_ver"
text="$(curl  "https://api.github.com/repos/gardener/gardener/contents/go.mod?ref=${gardener_ver}" --header "authorization: Bearer ${GH_TOKEN}" 2>/dev/null | jq -r ".content" | base64 --decode )"
echo "$text"
