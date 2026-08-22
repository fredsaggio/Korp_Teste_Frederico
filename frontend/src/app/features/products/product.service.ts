import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { CreateProductInput, Product } from './product';

@Injectable({ providedIn: 'root' })
export class ProductService {
  private readonly http = inject(HttpClient);
  private readonly endpoint = '/api/v1/products';

  list(): Observable<Product[]> {
    return this.http.get<Product[]>(this.endpoint);
  }

  create(input: CreateProductInput): Observable<Product> {
    return this.http.post<Product>(this.endpoint, input);
  }
}
