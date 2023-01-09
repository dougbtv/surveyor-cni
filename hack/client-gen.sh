#!/bin/bash
./vendor/k8s.io/code-generator/generate-groups.sh "client,informer,lister" \
  github.com/dougbtv/surveyor-cni/pkg/client \
  github.com/dougbtv/surveyor-cni/pkg/apis \
  k8s.cni.cncf.io:v1 \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt
