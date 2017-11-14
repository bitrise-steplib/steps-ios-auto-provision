#!/bin/bash

set -e
THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export BUNDLE_GEMFILE="$THIS_SCRIPT_DIR/Gemfile"

set +e
out=$(bundle install)
if [ $? != 0 ]; then
    echo "bundle install failed"
    echo $out
    exit 1
fi
set -e

bundle exec ruby "$THIS_SCRIPT_DIR/step.rb"