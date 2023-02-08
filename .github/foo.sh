#!/bin/bash

set -e

gardener_ver="$(sed -E -n 's/\s*github.com\/gardener\/gardener\s+(\S+)/\1/p' go.mod)"
echo "$gardener_ver"
text="$(curl  "https://api.github.com/repos/gardener/gardener/contents/go.mod?ref=v1.60.0" --header "authorization: Bearer ${GH_TOKEN}" )"
echo "$text"