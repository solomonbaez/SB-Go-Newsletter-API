set -x
set -eo pipefail

# dependencies
if ! [ -x "$(command -v psql)" ]; then
    echo >&2 "ERROR: psql is not installed"
    exit 1
fi

if ! [ -x "$(command -v go)" ]; then
    echo >&2 "ERROR: Go is not installed"
    exit 1
fi

if ! [ -x "$(command -v migrate)" ]; then
    echo >&2 "INFO: Go migrate is not installed - installing"

    go install \
        -tags 'postgres' \
        github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# migrations
MIGRATIONS_DIR="./db/migrations"
if [ ! -d "${MIGRATIONS_DIR}" ]; then
    mkdir -p "${MIGRATIONS_DIR}"
    migrate create -ext sql -dir "${MIGRATIONS_DIR}" -seq create_subscriptions_table
fi

# configurations
DB_USER="${POSTGRES_USER:=postgres}"
DB_PASS="${POSTGRES_PASS:=password}"
DB_NAME="${POSTGRES_NAME:=newsletter}"
DB_HOST="${POSTGRES_HOST:=localhost}"
DB_PORT="${POSTGRES_PORT:=5432}"
SSLMODE="disable" 

# check docker status
if [[ -z "${SKIP_DOCKER}" ]]
then
    docker run \
        --name newsletter \
        -e POSTGRES_USER="${DB_USER}" \
        -e POSTGRES_PASSWORD="${DB_PASS}" \
        -e POSTGRES_DB="${DB_NAME}" \
        -p "${DB_PORT}":5432 \
        -d postgres  
fi

until psql -h "${DB_HOST}" -U "${DB_USER}" -p "${DB_PORT}" -d "postgres" -c '\q'; do
    >&2 echo "Postgres is still unavailable - sleeping"
    sleep 1
done

>&2 echo "INFO: postgres is running on port ${DB_PORT} - attempting to apply migrations"

# environment
DB_URL="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${SSLMODE}"
export DB_URL
export DB_PASS

# apply migrations
migrate -source "file://${MIGRATIONS_DIR}" -database "${DB_URL}" up
>&2 echo "INFO: successfully applied migrations to postgres database ${DB_NAME}"
