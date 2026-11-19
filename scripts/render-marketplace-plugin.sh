#!/usr/bin/env bash

set -euo pipefail

readonly APP_VERSION="${APP_VERSION:?Set APP_VERSION to the released application version}"
readonly AMD64_SHA256="${AMD64_SHA256:?Set AMD64_SHA256 to the release checksum}"
readonly ARM64_SHA256="${ARM64_SHA256:?Set ARM64_SHA256 to the release checksum}"
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
readonly TEMPLATE_DIR="${REPO_ROOT}/distro/omarchy-plugin"
readonly OUTPUT_DIR="${REPO_ROOT}/marketplace"

[[ "${APP_VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]
[[ "${AMD64_SHA256}" =~ ^[[:xdigit:]]{64}$ ]]
[[ "${ARM64_SHA256}" =~ ^[[:xdigit:]]{64}$ ]]

mkdir -p "${OUTPUT_DIR}/assets"
cp "${TEMPLATE_DIR}/BarWidget.qml" "${OUTPUT_DIR}/BarWidget.qml"
cp "${TEMPLATE_DIR}/Version.js" "${OUTPUT_DIR}/Version.js"
cp "${TEMPLATE_DIR}/install-omarchy.sh" "${OUTPUT_DIR}/install-omarchy.sh"
cp "${TEMPLATE_DIR}/uninstall-app.sh" "${OUTPUT_DIR}/uninstall-app.sh"
cp "${TEMPLATE_DIR}/README.md" "${OUTPUT_DIR}/README.md"
cp "${REPO_ROOT}/assets/io.github.spalencsar.OmarchyFiles.desktop" "${OUTPUT_DIR}/assets/"
cp "${REPO_ROOT}/assets/io.github.spalencsar.OmarchyFiles.svg" "${OUTPUT_DIR}/assets/"
cp "${REPO_ROOT}/docs/screenshots/main-window.png" "${REPO_ROOT}/preview.png"
cp "${TEMPLATE_DIR}/manifest.json" "${REPO_ROOT}/manifest.json"

for file in \
  "${OUTPUT_DIR}/BarWidget.qml" \
  "${OUTPUT_DIR}/install-omarchy.sh" \
  "${OUTPUT_DIR}/README.md" \
  "${REPO_ROOT}/manifest.json"; do
  sed -i \
    -e "s/@APP_VERSION@/${APP_VERSION}/g" \
    -e "s/@AMD64_SHA256@/${AMD64_SHA256}/g" \
    -e "s/@ARM64_SHA256@/${ARM64_SHA256}/g" \
    "${file}"
done

sed -i 's|"BarWidget.qml"|"marketplace/BarWidget.qml"|' \
  "${REPO_ROOT}/manifest.json"
chmod 0755 "${OUTPUT_DIR}/install-omarchy.sh" "${OUTPUT_DIR}/uninstall-app.sh"

bash -n "${OUTPUT_DIR}/install-omarchy.sh" "${OUTPUT_DIR}/uninstall-app.sh"
jq -e . "${REPO_ROOT}/manifest.json" >/dev/null
if grep -R '@APP_VERSION@\|@AMD64_SHA256@\|@ARM64_SHA256@' \
  "${OUTPUT_DIR}/BarWidget.qml" \
  "${OUTPUT_DIR}/install-omarchy.sh" \
  "${OUTPUT_DIR}/README.md" \
  "${REPO_ROOT}/manifest.json"; then
  echo 'Unresolved Marketplace plugin template placeholder' >&2
  exit 1
fi
