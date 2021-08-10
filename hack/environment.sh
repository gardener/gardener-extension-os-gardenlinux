#!/usr/bin/env bash
#
# Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

set -e

get_cd_registry () {
  if [ -n "$registry" ]; then
    echo $registry
  else
    info "For creating and uploading a component descriptor just set the env variable registry to the repository of your oci registry e.g. eu.gcr.io/.../cnudie/gardener/development"
    exit 1
  fi
}

get_cd_component_name () {
  echo "github.com/gardener/gardener-extension-os-gardenlinux"
}

get_image_registry () {
  echo "eu.gcr.io/gardener-project/gardener/extensions"
}