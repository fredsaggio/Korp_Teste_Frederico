package main

import (
	"korp/faturamento-service/internal/db"
	"korp/faturamento-service/internal/invoices"
	"korp/faturamento-service/internal/server"
	"korp/faturamento-service/internal/stock"
)

func buildHandlers(database db.DB, stockClient stock.Client) server.Handlers {
	invoiceStore := invoices.NewInvoiceStore(database)
	invoiceService := invoices.NewInvoiceService(invoiceStore)
	invoiceHandler := invoices.NewInvoiceHandler(invoiceService)
	invoiceClosingService := invoices.NewInvoiceClosingService(invoiceStore, stockClient)
	invoiceClosingHandler := invoices.NewInvoiceClosingHandler(invoiceClosingService)

	return server.Handlers{
		InvoiceHandler:        invoiceHandler,
		InvoiceClosingHandler: invoiceClosingHandler,
	}
}
