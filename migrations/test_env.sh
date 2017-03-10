#!/bin/bash
set -e -o pipefail

function abort() {
  echo $@
  exit 1
}

docker-compose create
docker-compose start

port=$(docker-compose port database_test 5432 | cut -d ":" -f 2)

if [[ -z $port ]]; then
  abort "unable to get port of database"
fi

echo "export TEST_WITH_DB=true"
echo "export TEST_DATABSE_URL=postgres://postgres@127.0.0.1:${port}/dgtk_migrations?sslmode=disable"
