package openapi

//go:generate go get github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate types,client -o ./openapi.go -package openapi ./openapi.yaml
//go:generate go mod tidy
