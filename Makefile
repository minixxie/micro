GOLANG_IMG=minixxie/golang:1.21.0

golang:
	docker run --rm -it --net=host -v "${PWD}:/go/src/app" -w "/go/src/app" -e ENV=local "${GOLANG_IMG}" bash
