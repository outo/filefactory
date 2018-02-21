#!/usr/bin/env bash

while [ $? -eq 0 ]; do ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress -p; done
