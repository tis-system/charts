#!/usr/bin/env bash

# This requires gnu-sed to be installed.
# TODO(dio): Port to Go.

set -e

readonly ISTIO=$1
readonly VERSION=$2

export VERSION

find ./charts/${ISTIO}/${VERSION} -name 'Chart.yaml'  -type f -exec yq -i '.version=env(VERSION)' {} \;
sed -i '/\$whv := dict/,8d' charts/${ISTIO}/${VERSION}/istio-control/istio-discovery/templates/revision-tags.yaml
sed -e '4i {{- $whv := dict "revision" .Values.revision "injectionPath" .Values.istiodRemote.injectionPath "injectionURL" .Values.istiodRemote.injectionURL "namespace" .Release.Namespace }}' -i charts/${ISTIO}/${VERSION}/istio-control/istio-discovery/templates/revision-tags.yaml
sed -i '/\$whv := dict/,7d' charts/${ISTIO}/${VERSION}/istio-control/istio-discovery/templates/mutatingwebhook.yaml
sed -e '3i {{- $whv := dict "revision" .Values.revision "injectionPath" .Values.istiodRemote.injectionPath "injectionURL" .Values.istiodRemote.injectionURL "namespace" .Release.Namespace }}'  -i charts/${ISTIO}/${VERSION}/istio-control/istio-discovery/templates/mutatingwebhook.yaml
