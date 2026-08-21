package main

import (
	"korp/estoque-service/internal/db"
	"korp/estoque-service/internal/products"
	"korp/estoque-service/internal/server"
	"korp/estoque-service/internal/stock"
)

func buildHandlers(database db.DB) server.Handlers {
	productStore := products.NewProductStore(database)
	productService := products.NewProductService(productStore)
	productHandler := products.NewProductHandler(productService)
	debitStore := stock.NewDebitStore(database)
	debitService := stock.NewDebitService(debitStore)
	debitHandler := stock.NewDebitHandler(debitService)

	return server.Handlers{
		ProductHandler: productHandler,
		DebitHandler:   debitHandler,
	}
}
