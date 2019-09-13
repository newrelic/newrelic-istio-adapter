#!/usr/bin/env bash
#
# Copyright 2019 New Relic Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Bump project version in preperation for release.

set -e

readonly BIN="$( basename "$0" )"
readonly BIN_DIR="$( cd "$(dirname "$0")" ; pwd -P )"
readonly VERSION_FILE="${BIN_DIR}/../VERSION"
readonly CHANGELOG_FILE="${BIN_DIR}/../CHANGELOG.md"
readonly HELM_CHART_MANIFEST="${BIN_DIR}/../helm-charts/Chart.yaml"

# Matches the expected semver spec: https://semver.org/
readonly SEMVER_RE="^(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"

# Display CLI help dialogue.
usage() {
    cat <<HELP_USAGE
usage: ${BIN} VERSION

arguments:
    VERSION      Semver format version to set project to.
HELP_USAGE
}


validate_version() {
    local version="$1"
    if ! [[ "$version" =~ $SEMVER_RE ]]
    then
        echo "invalid semver: $version"
        usage
        exit 2
    fi
}

main() {
    local version

    if [ $# -eq 0 ]
    then
         echo "missing VERSION"
         usage >&2
         exit 1
    fi

    if [ $# -gt 1 ]
    then
         echo "unknown arguments: ${*:2}"
         usage >&2
         exit 1
    fi

    version="$1"
    validate_version "$version"

    # Add version release section to changelog.
    sed -i.bkup -e "/^## Unreleased.*/a\\
\\
## $version\\
" "$CHANGELOG_FILE" && rm "${CHANGELOG_FILE}.bkup"

    # Update app version in helm chart
    sed -i.bkup -e "/^appVersion.*/s/.*/appVersion: \"$version\"/" "$HELM_CHART_MANIFEST" \
        && rm "${HELM_CHART_MANIFEST}.bkup"

    # Update the project version file.
    echo "$version" > "$VERSION_FILE"
}

main "$@"
