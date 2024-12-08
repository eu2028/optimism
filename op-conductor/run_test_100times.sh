#!/bin/bash

set -e

for i in {1..10000}; do
  echo "======================="
  echo "Running iteration $i"
  # if ! gotestsum -- -run 'TestControlLoop' ./conductor/... --count=1 --timeout=5s -race; then
  #   echo "Test failed"
  #   exit 1
  # fi
  # if ! go test -v ./conductor/... -run ^TestScenario2$ -race -count=1; then
  #   echo "Test failed"
  #   exit 1
  # fi

  if ! go test -v ./conductor/... -race -count=1; then
    echo "Test failed"
    exit 1
  fi
done
