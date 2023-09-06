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
POST http://golang_grpc.local.test.com/v1/SecondService/Echo
Content-Type: application/json; charset=utf8

{
	"name": "Simon"
}
EOF

cd "$scriptPath"/..

SCRIPT="$0" \
BODY_FILE="$tmpfile" \
NET=host \
make curl
