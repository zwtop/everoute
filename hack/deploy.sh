#!/usr/bin/env bash

# Copyright 2021 The Everoute Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o pipefail
set -o nounset

local_path=$(dirname "$(readlink -f "${0}")")
kubectl apply -f "${local_path}"/../deploy/everoute.yaml

for i in {1..100}; do
  kubectl get po -Aowide
  kubectl describe po -l app=everoute -n kube-system || true
  kubectl logs -l component=everoute-agent -c init-agent -n kube-system || true
  kubectl logs -l component=everoute-agent -c everoute-agent -n kube-system || true
  kubectl logs -l component=everoute-controller -n kube-system || true
  sleep 2
done

kubectl wait po -n kube-system --for=condition=Ready -l app=everoute --timeout=3m

echo "========================================================="
echo " "
echo "Installation is complete for everoute !"
echo " "
echo "========================================================="
