import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./features/tasks/task-list/task-list').then((m) => m.TaskList),
    title: 'Tasks',
  },
];
