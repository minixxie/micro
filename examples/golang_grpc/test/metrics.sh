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
GET http://local_golang_grpc:80/metrics
EOF

cd "$scriptPath"/..

SCRIPT="$0" \
BODY_FILE="$tmpfile" \
NET=local_network \
make curl
