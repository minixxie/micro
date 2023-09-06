#!/bin/bash

scriptPath=$(cd $(dirname "$0") && pwd)
tmpfile=$(mktemp /tmp/XXXXXX)
function cleanup {
   rm -f "$tmpfile"
}
trap cleanup EXIT

function uuid() {
    if [ $(which uuidgen) ]; then
        uuidgen | tr '[A-Z]' '[a-z]'
    else
        uuid
    fi
}

cat <<EOF > "$tmpfile"
POST http://local_golang_grpc/v1/FirstService/TryDataTypes
Content-Type: application/json; charset=utf8

{
	"time": "2017-01-01T09:30:59Z"
}
EOF

cd "$scriptPath"/..

SCRIPT="$0" \
BODY_FILE="$tmpfile" \
NET=local_network \
make curl
