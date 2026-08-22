import { DatePipe } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import {
  ChangeDetectionStrategy,
  Component,
  DestroyRef,
  OnInit,
  inject,
  signal,
} from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar } from '@angular/material/snack-bar';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { finalize, forkJoin } from 'rxjs';
import { Product } from '../products/product';
import { ProductService } from '../products/product.service';
import { Invoice } from './invoice';
import { InvoiceService } from './invoice.service';

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [DatePipe, MatButtonModule, MatCardModule, MatProgressSpinnerModule, RouterLink],
  selector: 'app-invoice-detail',
  styleUrl: './invoice-detail.scss',
  templateUrl: './invoice-detail.html',
})
export class InvoiceDetail implements OnInit {
  private readonly destroyRef = inject(DestroyRef);
  private readonly invoiceService = inject(InvoiceService);
  private readonly productService = inject(ProductService);
  private readonly route = inject(ActivatedRoute);
  private readonly snackBar = inject(MatSnackBar);

  private readonly number = Number(this.route.snapshot.paramMap.get('number'));

  protected readonly invoice = signal<Invoice | null>(null);
  protected readonly products = signal<Product[]>([]);
  protected readonly loading = signal(true);
  protected readonly processing = signal(false);
  protected readonly errorMessage = signal<string | null>(null);

  ngOnInit(): void {
    this.loadInvoice();
  }

  protected loadInvoice(): void {
    if (!Number.isInteger(this.number) || this.number <= 0) {
      this.loading.set(false);
      this.errorMessage.set('A numeração da nota é inválida.');
      return;
    }

    this.loading.set(true);
    this.errorMessage.set(null);
    forkJoin({
      invoice: this.invoiceService.get(this.number),
      products: this.productService.list(),
    })
      .pipe(
        takeUntilDestroyed(this.destroyRef),
        finalize(() => this.loading.set(false)),
      )
      .subscribe({
        next: ({ invoice, products }) => {
          this.invoice.set(invoice);
          this.products.set(products);
        },
        error: (error: unknown) => this.errorMessage.set(this.getErrorMessage(error)),
      });
  }

  protected closeAndPrint(): void {
    const invoice = this.invoice();
    if (!invoice || invoice.status !== 'OPEN' || this.processing()) {
      return;
    }

    this.processing.set(true);
    this.invoiceService
      .close(invoice.number)
      .pipe(
        takeUntilDestroyed(this.destroyRef),
        finalize(() => this.processing.set(false)),
      )
      .subscribe({
        next: (closedInvoice) => {
          this.invoice.set(closedInvoice);
          this.snackBar.open('Nota fechada e estoque atualizado.', 'Fechar', { duration: 3000 });
          window.setTimeout(() => window.print());
        },
        error: (error: unknown) => {
          this.snackBar.open(this.getErrorMessage(error), 'Fechar', { duration: 5000 });
        },
      });
  }

  protected productLabel(productId: number): string {
    const product = this.products().find((item) => item.id === productId);
    return product ? `${product.code} — ${product.description}` : `Produto #${productId}`;
  }

  private getErrorMessage(error: unknown): string {
    if (
      error instanceof HttpErrorResponse &&
      typeof error.error === 'object' &&
      error.error !== null &&
      typeof error.error.error === 'string'
    ) {
      return error.error.error;
    }
    return 'Não foi possível concluir a operação. Tente novamente.';
  }
}
