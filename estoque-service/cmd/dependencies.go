package main

import (
	"korp/estoque-service/internal/db"
	"korp/estoque-service/internal/products"
	"korp/estoque-service/internal/server"
)

func buildHandlers(database db.DB) server.Handlers {
	productStore := products.NewProductStore(database)
	productService := products.NewProductService(productStore)
	productHandler := products.NewProductHandler(productService)

	return server.Handlers{
		ProductHandler: productHandler,
	}
}
