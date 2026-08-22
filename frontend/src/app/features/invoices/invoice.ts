export type InvoiceStatus = 'OPEN' | 'CLOSED';

export interface InvoiceItem {
  product_id: number;
  quantity: number;
}

export interface Invoice {
  number: number;
  status: InvoiceStatus;
  items: InvoiceItem[];
  created_at: string;
  updated_at: string;
  closed_at: string | null;
}

export interface CreateInvoiceInput {
  items: InvoiceItem[];
}
