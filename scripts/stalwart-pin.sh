#!/bin/sh
set -eu
pinned=$(grep -o 'stalwartlabs/stalwart:v[^" ]*' docker/docker-compose.yml | head -n 1 | cut -d: -f2)
latest=$(curl -fsSL --max-time 30 https://api.github.com/repos/stalwartlabs/stalwart/releases/latest | grep -o '"tag_name": *"[^"]*"' | head -n 1 | cut -d'"' -f4)
echo "pinned: $pinned latest: $latest"
if [ -z "$pinned" ] || [ -z "$latest" ]; then
  echo "could not determine pinned or latest stalwart version" >&2
  exit 2
fi
if [ "$latest" != "$pinned" ]; then
  echo "upstream moved to $latest (pinned $pinned); see docs/JMAP_WATCH.md" >&2
  exit 1
fi
