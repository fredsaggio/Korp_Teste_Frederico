import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { CreateInvoiceInput, Invoice } from './invoice';
import { InvoiceService } from './invoice.service';

describe('InvoiceService', () => {
  let service: InvoiceService;
  let http: HttpTestingController;

  const invoice: Invoice = {
    number: 1,
    status: 'OPEN',
    items: [{ product_id: 10, quantity: 2 }],
    created_at: '2026-08-21T12:00:00Z',
    updated_at: '2026-08-21T12:00:00Z',
    closed_at: null,
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [InvoiceService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(InvoiceService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('should list invoices', () => {
    service.list().subscribe((result) => expect(result).toEqual([invoice]));

    const request = http.expectOne('/api/v1/invoices');
    expect(request.request.method).toBe('GET');
    request.flush([invoice]);
  });

  it('should create an invoice', () => {
    const input: CreateInvoiceInput = { items: invoice.items };

    service.create(input).subscribe((result) => expect(result).toEqual(invoice));

    const request = http.expectOne('/api/v1/invoices');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(input);
    request.flush(invoice);
  });
});
