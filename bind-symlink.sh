#!/bin/bash

bind-symlink() {
    gravity_adapter_project_path=$1
    binary_name=$2

    dst_sl="$(pwd)/binaries/$binary_name"

    current=$(pwd)
    cd $gravity_adapter_project_path

    origin_sl="$PWD/target/deploy/solana_gravity_adaptor.so"

    sudo ln -s -f $origin_sl $dst_sl

    cd $current
}

# bind nebula to symlink
bind-symlink "../solana-adapter/src/gravity-core-adapter" "nebula.so"