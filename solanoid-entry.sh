#!/bin/bash

if ! command -v spl-token &> /dev/null
then
    echo "spl-token is not installed"
    echo "installing Solana CLI binaries..."
    sh -c "$(curl -sSfL https://release.solana.com/v1.7.3/install)"
fi

echo 'Starting Polygon->Solana $GTON MVP..'

./solanoid_macos