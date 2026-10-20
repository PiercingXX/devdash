#!/bin/bash

goreleaser release --clean && rm -rf dist/
