import { HttpErrorResponse } from '@angular/common/http';
import { DatePipe } from '@angular/common';
import {
  ChangeDetectionStrategy,
  Component,
  DestroyRef,
  OnInit,
  inject,
  signal,
} from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormArray, FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatTableModule } from '@angular/material/table';
import { finalize, forkJoin } from 'rxjs';
import { Product } from '../products/product';
import { ProductService } from '../products/product.service';
import { Invoice } from './invoice';
import { InvoiceService } from './invoice.service';

type InvoiceItemForm = FormGroup<{
  productId: FormControl<number | null>;
  quantity: FormControl<number>;
}>;

@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    DatePipe,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatProgressSpinnerModule,
    MatSelectModule,
    MatTableModule,
    ReactiveFormsModule,
  ],
  selector: 'app-invoices',
  styleUrl: './invoices.scss',
  templateUrl: './invoices.html',
})
export class Invoices implements OnInit {
  private readonly destroyRef = inject(DestroyRef);
  private readonly invoiceService = inject(InvoiceService);
  private readonly productService = inject(ProductService);
  private readonly snackBar = inject(MatSnackBar);

  protected readonly displayedColumns = ['number', 'createdAt', 'items', 'status'];
  protected readonly products = signal<Product[]>([]);
  protected readonly invoices = signal<Invoice[]>([]);
  protected readonly loading = signal(true);
  protected readonly saving = signal(false);
  protected readonly errorMessage = signal<string | null>(null);

  protected readonly form = new FormGroup({
    items: new FormArray<InvoiceItemForm>([this.createItemForm()]),
  });

  protected get items(): FormArray<InvoiceItemForm> {
    return this.form.controls.items;
  }

  ngOnInit(): void {
    this.loadData();
  }

  protected loadData(): void {
    this.loading.set(true);
    this.errorMessage.set(null);

    forkJoin({
      products: this.productService.list(),
      invoices: this.invoiceService.list(),
    })
      .pipe(
        takeUntilDestroyed(this.destroyRef),
        finalize(() => this.loading.set(false)),
      )
      .subscribe({
        next: ({ products, invoices }) => {
          this.products.set(products);
          this.invoices.set(invoices);
        },
        error: (error: unknown) => this.errorMessage.set(this.getErrorMessage(error)),
      });
  }

  protected addItem(): void {
    this.items.push(this.createItemForm());
  }

  protected removeItem(index: number): void {
    if (this.items.length > 1) {
      this.items.removeAt(index);
    }
  }

  protected isProductSelected(productId: number, currentIndex: number): boolean {
    return this.items.controls.some(
      (item, index) => index !== currentIndex && item.controls.productId.value === productId,
    );
  }

  protected submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const productIds = this.items.controls.map((item) => item.controls.productId.value);
    if (new Set(productIds).size !== productIds.length) {
      this.snackBar.open('Um produto não pode aparecer mais de uma vez na nota.', 'Fechar', {
        duration: 5000,
      });
      return;
    }

    const input = {
      items: this.items.controls.map((item) => ({
        product_id: item.controls.productId.value!,
        quantity: item.controls.quantity.value,
      })),
    };

    this.saving.set(true);
    this.invoiceService
      .create(input)
      .pipe(
        takeUntilDestroyed(this.destroyRef),
        finalize(() => this.saving.set(false)),
      )
      .subscribe({
        next: (invoice) => {
          this.invoices.update((invoices) => [...invoices, invoice]);
          this.resetForm();
          this.snackBar.open(`Nota fiscal nº ${invoice.number} criada com sucesso.`, 'Fechar', {
            duration: 3000,
          });
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

  protected statusLabel(status: Invoice['status']): string {
    return status === 'OPEN' ? 'Aberta' : 'Fechada';
  }

  private createItemForm(): InvoiceItemForm {
    return new FormGroup({
      productId: new FormControl<number | null>(null, Validators.required),
      quantity: new FormControl(1, {
        nonNullable: true,
        validators: [Validators.required, Validators.min(1)],
      }),
    });
  }

  private resetForm(): void {
    this.items.clear();
    this.items.push(this.createItemForm());
    this.form.markAsPristine();
    this.form.markAsUntouched();
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
