set -x
set -eo pipefail

# dependencies
if ! [ -x "$(command -v psql)" ]; then
    echo >&2 "ERROR: psql is not installed"
    exit 1
fi

if ! [ -x "${command -v migrate}"]; then
    echo >&2 "INFO: Go migrate is not installed - installing"

    go install \
        -tags 'postgres' \
        github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
fi

# configurations
DB_USER="${POSTGRES_USER:=postgres}"
DB_PASS="${POSTGRES_PASS:=password}"
DB_NAME="${POSTGRES_NAME:=newsletter}"
DB_PORT="${POSTGRES_PORT:=5432}"
DB_HOST="${POSTGRES_HOST:=localhost}"

MIGRATIONS_DIR="./db/migrations"
if [ ! -d "${MIGRATIONS_DIR}" ]; then
    mkdir -p "${MIGRATIONS_DIR}"
fi

# DEV - remove test container
docker stop newsletter
docker rm newsletter

docker run \
    --name newsletter \
    -e POSTGRES_USER=${DB_USER} \
    -e POSTGRES_PASSWORD=${DB_PASS} \
    -e POSTGRES_DB=${DB_NAME} \
    -p "${DB_PORT}":5432 \
    -d postgres  

export PG_PASS="${DB_PASS}"
until psql -h "${DB_HOST}" -U "${DB_USER}" -p "${DB_PORT}" -d "postgres" -c '\q'; do
    >&2 echo "Postgres is still unavailable - sleeping"
    sleep 1
done

migrate create -ext sql -dir ./db/migrations -seq create_subscriptions_table

>&2 echo "Success! Postgres is running on port ${DB_PORT}!"

# environment
DB_URL=postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}
export DATABASE_URL