import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { CreateInvoiceInput, Invoice } from './invoice';

@Injectable({ providedIn: 'root' })
export class InvoiceService {
  private readonly http = inject(HttpClient);
  private readonly endpoint = '/api/v1/invoices';

  list(): Observable<Invoice[]> {
    return this.http.get<Invoice[]>(this.endpoint);
  }

  create(input: CreateInvoiceInput): Observable<Invoice> {
    return this.http.post<Invoice>(this.endpoint, input);
  }

  get(number: number): Observable<Invoice> {
    return this.http.get<Invoice>(`${this.endpoint}/${number}`);
  }

  close(number: number): Observable<Invoice> {
    return this.http.post<Invoice>(`${this.endpoint}/${number}/close`, {});
  }
}
