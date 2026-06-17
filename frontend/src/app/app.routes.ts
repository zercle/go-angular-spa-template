import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./features/tasks/task-list/task-list').then((m) => m.TaskList),
    title: 'Tasks',
  },
  {
    path: '**',
    loadComponent: () =>
      import('./features/not-found/not-found').then((m) => m.NotFound),
    title: 'Page not found',
  },
];
