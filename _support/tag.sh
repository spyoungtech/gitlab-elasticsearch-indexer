set -e

main() {
  get_version
  tag_name="v${version}"
  git tag $TAG_OPTS -m "Version ${version}" -a ${tag_name}
  git show ${tag_name}
  cat <<'EOF'

  Remember to now push your tag, either to gitlab.com (for a
  normal release) or dev.gitlab.org (for a security release).
EOF
}

get_version() {
  v=$(sed 1q VERSION)
  if ! echo "${v}" | grep -q '^[0-9]\+\.[0-9]\+\.[0-9]\+$' ; then
    echo "Invalid VERSION: ${v}"
    exit 1
  fi
  version="${v}"
}

main
