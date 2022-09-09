package v2

//go:generate sh -c "go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config ./openapi/config.yaml ../../../docs/openapi/v2.yaml > openapi/openapi.gen.go"
//go:generate go fmt ./openapi/openapi.gen.go
