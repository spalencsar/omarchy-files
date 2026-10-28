#!/usr/bin/env bash

set -euo pipefail

readonly APP_VERSION="@APP_VERSION@"
readonly RELEASE_BASE="https://github.com/spalencsar/omarchy-files/releases/download/v${APP_VERSION}"
readonly INSTALL_PREFIX="${OMARCHY_FILES_INSTALL_PREFIX:-${HOME}/.local}"
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

case "$(uname -m)" in
  x86_64)
    readonly RELEASE_ARCH="amd64"
    readonly EXPECTED_SHA256="@AMD64_SHA256@"
    ;;
  aarch64|arm64)
    readonly RELEASE_ARCH="arm64"
    readonly EXPECTED_SHA256="@ARM64_SHA256@"
    ;;
  *)
    echo "Unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

if ! command -v omarchy >/dev/null 2>&1; then
  echo "Omarchy Files requires Omarchy." >&2
  exit 1
fi

cat <<EOF
Omarchy Files v${APP_VERSION}

This installs the signed release binary for the current user under:
  ${INSTALL_PREFIX}

It will not change your default file manager, enable a service, or select a
file chooser portal. The download is checked against a pinned SHA-256 hash.
EOF

printf "\nContinue? [y/N] "
read -r answer
case "${answer}" in
  y|Y|yes|YES) ;;
  *) echo "Installation cancelled."; exit 0 ;;
esac

download_dir="$(mktemp -d -t omarchy-files-install.XXXXXX)"
trap 'rm -rf -- "${download_dir}"' EXIT

archive="omarchy-files-linux-${RELEASE_ARCH}.gz"
echo
echo "Downloading ${archive}..."
curl --fail --location --proto '=https' --tlsv1.2 \
  "${RELEASE_BASE}/${archive}" \
  --output "${download_dir}/${archive}"

printf '%s  %s\n' "${EXPECTED_SHA256}" "${download_dir}/${archive}" \
  | sha256sum --check --status
echo "Checksum verified."

install -d -- \
  "${INSTALL_PREFIX}/bin" \
  "${INSTALL_PREFIX}/share/applications" \
  "${INSTALL_PREFIX}/share/icons/hicolor/scalable/apps"

gzip --decompress --stdout "${download_dir}/${archive}" \
  > "${download_dir}/omarchy-files"
install -m 0755 "${download_dir}/omarchy-files" \
  "${INSTALL_PREFIX}/bin/omarchy-files"
install -m 0644 "${SCRIPT_DIR}/assets/io.github.spalencsar.OmarchyFiles.svg" \
  "${INSTALL_PREFIX}/share/icons/hicolor/scalable/apps/io.github.spalencsar.OmarchyFiles.svg"
install -m 0644 "${SCRIPT_DIR}/assets/io.github.spalencsar.OmarchyFiles.desktop" \
  "${INSTALL_PREFIX}/share/applications/io.github.spalencsar.OmarchyFiles.desktop"

if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database "${INSTALL_PREFIX}/share/applications" >/dev/null 2>&1 || true
fi

if command -v gtk-update-icon-cache >/dev/null 2>&1; then
  gtk-update-icon-cache -f -t "${INSTALL_PREFIX}/share/icons/hicolor" >/dev/null 2>&1 || true
fi

echo
echo "Omarchy Files v${APP_VERSION} is installed for the current user."
echo "Click the bar widget again to open it."
echo "Optional integrations remain disabled until you explicitly enable them."
