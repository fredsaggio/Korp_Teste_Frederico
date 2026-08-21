package main

import (
	"korp/faturamento-service/internal/db"
	"korp/faturamento-service/internal/invoices"
	"korp/faturamento-service/internal/server"
)

func buildHandlers(database db.DB) server.Handlers {
	invoiceStore := invoices.NewInvoiceStore(database)
	invoiceService := invoices.NewInvoiceService(invoiceStore)
	invoiceHandler := invoices.NewInvoiceHandler(invoiceService)

	return server.Handlers{
		InvoiceHandler: invoiceHandler,
	}
}
