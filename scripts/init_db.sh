set -x
set -eo pipefail

# configurations
DB_USER="${POSTGRES_USER:=postgres}"
DB_PASS="${POSTGRES_PASS:=password}"
DB_NAME="${POSTGRES_NAME:=newsletter}"
DB_PORT="${POSTGRES_PORT:=5432}"
DB_HOST="${POSTGRES_HOST:=localhost}"

docker run \
    --name newsletter \
    -e POSTGRES_USER=${DB_USER} \
    -e POSTGRES_PASSWORD=${DB_PASS} \
    -e POSTGRES_DB=${DB_NAME} \
    -p "${DB_PORT}":5432 \
    -d postgres  