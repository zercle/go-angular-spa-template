import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { TaskStore } from '../../../core/task-store';
import { TaskForm, TaskFormValue } from '../task-form/task-form';

/** The tasks page: an add form plus the list with toggle/edit/delete. */
@Component({
  selector: 'app-task-list',
  imports: [MatButtonModule, MatCheckboxModule, MatIconModule, MatProgressBarModule, TaskForm],
  templateUrl: './task-list.html',
  styleUrl: './task-list.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class TaskList {
  readonly store = inject(TaskStore);

  /** Id of the task currently being edited inline, if any. */
  readonly editingId = signal<number | null>(null);

  constructor() {
    void this.store.load();
  }

  onCreate(value: TaskFormValue): void {
    void this.store.add({ title: value.title });
  }

  onUpdate(id: number, value: TaskFormValue): void {
    void this.store.update(id, value).then(() => this.editingId.set(null));
  }
}
