#!/bin/bash

set -e

cat "${REPO_DIR}/go.mod"
#gardener_ver="$(sed -E -n 's/\s*github.com\/gardener\/gardener\s+(\S+)/\1/p' "${REPO_DIR}/go.mod" )"
#echo "$gardener_ver"
#text="$(curl  "https://api.github.com/repos/gardener/gardener/contents/go.mod?ref=v1.60.0" --header "authorization: Bearer ${GH_TOKEN}" )"
#echo "$text"