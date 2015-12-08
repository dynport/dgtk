#!/bin/sh
set -e

if [[ -z $GIT_COMMIT ]]; then
  GIT_COMMIT=$(git rev-parse HEAD)
fi

HISTORY_LIMIT=${HISTORY_LIMIT:-50}
GIT_HISTORY=""
for name in $(git rev-list $GIT_COMMIT -n $HISTORY_LIMIT); do
	n=$(echo $name | cut -b 1-12)
	if [[ -n $GIT_HISTORY ]]; then
		GIT_HISTORY=${GIT_HISTORY},
	fi
	GIT_HISTORY=${GIT_HISTORY}'"'$n'"'
done

if [[ -n $GIT_HISTORY ]]; then
	GIT_HISTORY='['$GIT_HISTORY']'
fi

BUILT_AT="$(TZ=UTZ date +"%Y-%m-%dT%H:%M:%SZ")"

build=$(cat <<EOF
{
  "built_at":    "$BUILT_AT",
  "git_history": $GIT_HISTORY,
  "git_sha":     "$GIT_COMMIT",
  "changes":     $(if git status --porcelain > /dev/null 2>&1; then echo true; else false; fi),
  "hostname":    "$(hostname)",
  "user":        "$(whoami)"
}
EOF
)

echo $build | jq -c -r '.'
