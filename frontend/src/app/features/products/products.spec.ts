import { TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { Product } from './product';
import { ProductService } from './product.service';
import { Products } from './products';

describe('Products', () => {
  it('should load and render products on initialization', async () => {
    const product: Product = {
      id: 1,
      code: 'PROD-001',
      description: 'Produto um',
      balance: 10,
      created_at: '2026-08-21T12:00:00Z',
      updated_at: '2026-08-21T12:00:00Z',
    };
    const service = {
      list: () => of([product]),
      create: () => of(product),
    };
    await TestBed.configureTestingModule({
      imports: [Products],
      providers: [{ provide: ProductService, useValue: service }],
    }).compileComponents();

    const fixture = TestBed.createComponent(Products);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    const content = fixture.nativeElement.textContent as string;
    expect(content).toContain('PROD-001');
    expect(content).toContain('Produto um');
    expect(content).toContain('10');
  });
});
