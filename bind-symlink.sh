#!/bin/bash

bind-symlink() {
    gravity_adapter_project_path=$1
    symlink_binary_name=$2
    target_release_binary=$3

    dst_sl="$(pwd)/binaries/$symlink_binary_name"

    current=$(pwd)
    cd $gravity_adapter_project_path

    origin_sl="$PWD/$target_release_binary"

    sudo ln -s -f $origin_sl $dst_sl

    cd $current
}

bind-symlink "../solana-adapter/src/gravity-core-adapter" "gravity.so" "gravity/target/deploy/solana_gravity_contract.so"
bind-symlink "../solana-adapter/src/gravity-core-adapter" "nebula.so" "nebula/target/deploy/solana_nebula_contract.so"
bind-symlink "../solana-adapter/src/gravity-core-adapter" "ibport.so" "ibport/target/deploy/solana_ibport_contract.so"
bind-symlink "../solana-adapter/src/gravity-core-adapter" "luport.so" "luport/target/deploy/solana_luport_contract.so"

echo "build symlinks for binaries"
ls -la binaries/