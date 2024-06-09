package graph

import "graphql-sample/graph/services"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// 依存性注入
type Resolver struct {
	Srv services.Services
}
