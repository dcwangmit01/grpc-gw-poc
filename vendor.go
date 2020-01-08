// +build vendor

package main

// This file exists to trick "go mod vendor" to include "main" packages.
// It is not expected to build, the build tag above is only to prevent this
// file from being included in builds.
//
// borrowed from: https://github.com/dexidp/dex/blob/master/vendor.go

import (
	_ "github.com/cloudflare/cfssl/cmd/cfssl"
	_ "github.com/cloudflare/cfssl/cmd/cfssljson"
	_ "github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs"
	_ "github.com/gogo/protobuf/proto"
	_ "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"
	_ "github.com/jteeuwen/go-bindata/go-bindata"
	_ "github.com/onsi/ginkgo/ginkgo"
	_ "github.com/sugyan/ttyrec2gif"
	_ "golang.org/x/tools/cmd/goimports"
    _ "github.com/golang/protobuf/protoc-gen-go"
    _ "golang.org/x/lint/golint"
)

func main() {}
