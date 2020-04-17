#!/bin/bash

set -eu

version=${version-v4.0.0-rc}
old_version=${old_version-v4.0.0-beta.2}
name=test_upgrade
topo=./topo/upgrade.yaml

yes | tiup-cluster deploy $name $old_version $topo -i ~/.ssh/id_rsa

yes | tiup-cluster start $name

yes | tiup-cluster upgrade $name $version

yes | tiup-cluster destroy $name



