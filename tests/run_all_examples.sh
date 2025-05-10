#!/bin/bash

# This script runs openscadgen on all example config.toml files with the -ow flag.
# Usage: ./run_all_examples.sh

# Find all config.toml files under examples/ and process each one
find ../examples -type f -name "config.toml" | while read config; do
  echo "Processing $config"
  ../openscadgen -c "$config" -ow
  echo "---"
done

echo "All examples processed." 