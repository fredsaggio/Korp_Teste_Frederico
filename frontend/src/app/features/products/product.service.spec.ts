import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { CreateProductInput, Product } from './product';
import { ProductService } from './product.service';

describe('ProductService', () => {
  let service: ProductService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [ProductService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ProductService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('should list products', () => {
    const products: Product[] = [
      {
        id: 1,
        code: 'PROD-001',
        description: 'Produto um',
        balance: 10,
        created_at: '2026-08-21T12:00:00Z',
        updated_at: '2026-08-21T12:00:00Z',
      },
    ];

    service.list().subscribe((result) => expect(result).toEqual(products));

    const request = http.expectOne('/api/v1/products');
    expect(request.request.method).toBe('GET');
    request.flush(products);
  });

  it('should create a product', () => {
    const input: CreateProductInput = {
      code: 'PROD-001',
      description: 'Produto um',
      balance: 10,
    };
    const product: Product = {
      id: 1,
      ...input,
      created_at: '2026-08-21T12:00:00Z',
      updated_at: '2026-08-21T12:00:00Z',
    };

    service.create(input).subscribe((result) => expect(result).toEqual(product));

    const request = http.expectOne('/api/v1/products');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(input);
    request.flush(product);
  });
});
