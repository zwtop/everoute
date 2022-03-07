#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
set -o xtrace

# start ovs
/usr/share/openvswitch/scripts/ovs-ctl --no-ovs-vswitchd --system-id=random start
/usr/share/openvswitch/scripts/ovs-ctl --no-ovsdb-server --system-id=random start

eval $@
