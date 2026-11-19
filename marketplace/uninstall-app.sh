#!/usr/bin/env bash

set -euo pipefail

readonly INSTALL_PREFIX="${OMARCHY_FILES_INSTALL_PREFIX:-${HOME}/.local}"

cat <<EOF
This removes the user-local Omarchy Files application installed by this plugin.
It does not remove the bar plugin or your Omarchy Files settings.
EOF

printf "\nContinue? [y/N] "
read -r answer
case "${answer}" in
  y|Y|yes|YES) ;;
  *) echo "Removal cancelled."; exit 0 ;;
esac

"${INSTALL_PREFIX}/bin/omarchy-files" kill >/dev/null 2>&1 || true

rm -f -- \
  "${INSTALL_PREFIX}/bin/omarchy-files" \
  "${INSTALL_PREFIX}/share/applications/io.github.spalencsar.OmarchyFiles.desktop" \
  "${INSTALL_PREFIX}/share/icons/hicolor/scalable/apps/io.github.spalencsar.OmarchyFiles.svg"

if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database "${INSTALL_PREFIX}/share/applications" >/dev/null 2>&1 || true
fi

if command -v gtk-update-icon-cache >/dev/null 2>&1; then
  gtk-update-icon-cache -f -t "${INSTALL_PREFIX}/share/icons/hicolor" >/dev/null 2>&1 || true
fi

echo "The user-local Omarchy Files application was removed."
echo "Remove the bar plugin separately with:"
echo "  omarchy plugin remove io.github.spalencsar.omarchy-files"
