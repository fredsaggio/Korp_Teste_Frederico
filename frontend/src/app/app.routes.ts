import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('./features/home/home').then((component) => component.Home),
    title: 'Início | Korp Notas Fiscais',
  },
  { path: '**', redirectTo: '' },
];
