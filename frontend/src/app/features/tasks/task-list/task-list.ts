import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBar } from '@angular/material/snack-bar';

import { Task } from '../../../core/task';
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
  private readonly snackBar = inject(MatSnackBar);

  /** Id of the task currently being edited inline, if any. */
  readonly editingId = signal<string | null>(null);

  constructor() {
    void this.store.load();
  }

  async onCreate(value: TaskFormValue): Promise<void> {
    const ok = await this.store.add({ title: value.title });
    if (!ok) {
      this.notifyFailure();
    }
  }

  async onUpdate(id: string, value: TaskFormValue): Promise<void> {
    const ok = await this.store.update(id, value);
    if (ok) {
      this.editingId.set(null);
    } else {
      this.notifyFailure();
    }
  }

  async onToggle(task: Task): Promise<void> {
    const ok = await this.store.toggle(task);
    if (!ok) {
      this.notifyFailure();
    }
  }

  async onRemove(id: string): Promise<void> {
    const ok = await this.store.remove(id);
    if (!ok) {
      this.notifyFailure();
    }
  }

  /** Surfaces a mutation failure as a transient toast for screen + sighted users. */
  private notifyFailure(): void {
    this.snackBar.open(this.store.error() ?? 'Something went wrong.', 'Dismiss', {
      duration: 5000,
    });
  }
}
