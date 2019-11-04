#!/usr/bin/env bash

# Report coverage to sonarQube
echo "==> Report coverage to sonarQube..."

cp build/ci/sonar-project.properties .
sonar-scanner
rm -f coverage.out
rm -f sonar-project.properties
rm -fr .scannerwork/

exit 0