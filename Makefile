GOLANG_IMG := minixxie/golang:1.21.0

golang: local_network
	docker run --rm -it --net=local_network -v "${PWD}:/go/src/app" -w "/go/src/app" \
		"${GOLANG_IMG}" bash

gofmt:
	docker run --rm -t -v "${PWD}:/go/src/app" -w "/go/src/app" "${GOLANG_IMG}" gofmt -w .

local_network:
	docker network create -d bridge local_network 2> /dev/null || true