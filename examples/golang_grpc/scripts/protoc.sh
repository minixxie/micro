#!/bin/bash

go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-swagger
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
#go install github.com/micro/protobuf/protoc-gen-go
#go install github.com/micro/protobuf/proto
go install github.com/go-swagger/go-swagger/cmd/swagger
go install github.com/mwitkow/go-proto-validators/protoc-gen-govalidators

outputFolder=$(pwd)
cd pb
for pb in *.pb; do \
	echo "===> pb: $pb"
	cd $pb
	pkgFolders=$(find . -name "*.proto" | sed 's/[^/]*\.proto//' | sort | uniq)
	echo "pkgFolders: $pkgFolders"
	protoFileCount=$(find . -name "*.proto" | wc -l)
	if [ "$protoFileCount" -eq 0 ]; then echo >&2 "ERR: No *.proto files found."; exit 1; fi
	for pkgFolder in $pkgFolders
	do
		echo "pkgFolder: $pkgFolder"
		pwd
		find $pkgFolder -name "*.proto"
		find $pkgFolder -name "*.proto" | xargs \
		/usr/bin/protoc \
			-I. -I/usr/local/include -I$GOPATH/src \
			-I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis \
			-I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0 \
			-I$GOPATH/pkg/mod/github.com/mwitkow/go-proto-validators@v0.3.2 \
			--proto_path=. \
			--go_out "$outputFolder" --go_opt paths=source_relative \
			--go-grpc_out "$outputFolder" --go-grpc_opt paths=source_relative \
			--grpc-gateway_out=logtostderr=true:"$outputFolder" \
			--swagger_out=logtostderr=true:"$outputFolder" \
			--govalidators_out=Mgithub.com/gogo/protobuf/protobuf/google/protobuf/timestamp.proto=github.com/gogo/protobuf/types:"$outputFolder"
	done
	cd - 2>&1 > /dev/null
done

