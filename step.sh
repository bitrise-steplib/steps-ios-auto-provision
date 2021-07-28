#!/bin/bash

set -e
THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"


gem install bundler -v 2.2.24 --force

export BUNDLE_GEMFILE="$THIS_SCRIPT_DIR/Gemfile"

set +e
echo '$' "bundle install"
out=$(bundle install)
if [ $? != 0 ]; then
    echo "bundle install failed"
    echo $out
    exit 1
fi
set -e

echo '$' "bundle exec ruby "$THIS_SCRIPT_DIR/step.rb""
bundle exec ruby "$THIS_SCRIPT_DIR/step.rb"