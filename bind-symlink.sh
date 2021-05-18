#!/bin/bash


bind-symlink() {
    gravity_adapter_project_path=$1
    binary_name=$2
    sudo ln -s "$gravity_adapter_project_path/target/deploy/solana_gravity_adaptor.so" "$PWD/binaries/$binary_name"
}

# bind nebula to symlink
bind-symlink "../solana-adapter/src/gravity-core-adapter" "nebula.so"