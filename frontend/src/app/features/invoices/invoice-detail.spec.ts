import { convertToParamMap, provideRouter } from '@angular/router';
import { ActivatedRoute } from '@angular/router';
import { TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { ProductService } from '../products/product.service';
import { InvoiceService } from './invoice.service';
import { InvoiceDetail } from './invoice-detail';

describe('InvoiceDetail', () => {
  it('should render an open invoice and its products', async () => {
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
      get: () =>
        of({
          number: 1,
          status: 'OPEN' as const,
          items: [{ product_id: 10, quantity: 2 }],
          created_at: '2026-08-21T12:00:00Z',
          updated_at: '2026-08-21T12:00:00Z',
          closed_at: null,
        }),
      close: () => {
        throw new Error('not used');
      },
    };

    await TestBed.configureTestingModule({
      imports: [InvoiceDetail],
      providers: [
        provideRouter([]),
        { provide: ProductService, useValue: productService },
        { provide: InvoiceService, useValue: invoiceService },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ number: '1' }) } },
        },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(InvoiceDetail);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    const content = fixture.nativeElement.textContent as string;
    expect(content).toContain('Nº 1');
    expect(content).toContain('PROD-010');
    expect(content).toContain('Fechar e imprimir');
  });
});
