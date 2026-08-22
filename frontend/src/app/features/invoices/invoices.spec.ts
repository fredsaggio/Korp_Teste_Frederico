import { TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { ProductService } from '../products/product.service';
import { InvoiceService } from './invoice.service';
import { Invoices } from './invoices';

describe('Invoices', () => {
  it('should load and render products and invoices on initialization', async () => {
    const productService = {
      list: () =>
        of([
          {
            id: 10,
            code: 'PROD-010',
            description: 'Produto dez',
            balance: 8,
            created_at: '2026-08-21T12:00:00Z',
            updated_at: '2026-08-21T12:00:00Z',
          },
        ]),
    };
    const invoiceService = {
      list: () =>
        of([
          {
            number: 1,
            status: 'OPEN' as const,
            items: [{ product_id: 10, quantity: 2 }],
            created_at: '2026-08-21T12:00:00Z',
            updated_at: '2026-08-21T12:00:00Z',
            closed_at: null,
          },
        ]),
      create: () => {
        throw new Error('not used');
      },
    };
    await TestBed.configureTestingModule({
      imports: [Invoices],
      providers: [
        { provide: ProductService, useValue: productService },
        { provide: InvoiceService, useValue: invoiceService },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(Invoices);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    const content = fixture.nativeElement.textContent as string;
    expect(content).toContain('#1');
    expect(content).toContain('PROD-010');
    expect(content).toContain('Aberta');
  });
});
