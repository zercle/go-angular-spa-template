import { ChangeDetectionStrategy, Component, computed, effect, inject, input, output, viewChild } from '@angular/core';
import { FormBuilder, FormGroupDirective, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';

import { Task } from '../../../core/task';
import { TaskStore } from '../../../core/task-store';

/** Emitted when the user saves the form. */
export interface TaskFormValue {
  title: string;
  done: boolean;
}

/**
 * Reusable add/edit form for a task. When `task` is provided the form is in
 * edit mode (prefilled, shows the done toggle + Cancel); otherwise it adds.
 * Accessibility follows the modern-web-guidance forms guide.
 */
@Component({
  selector: 'app-task-form',
  imports: [ReactiveFormsModule, MatFormFieldModule, MatInputModule, MatButtonModule, MatCheckboxModule],
  templateUrl: './task-form.html',
  styleUrl: './task-form.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class TaskForm {
  /** When set, the form edits this task; when null, it creates a new one. */
  readonly task = input<Task | null>(null);

  /** Emits the form value on a valid submit. */
  readonly save = output<TaskFormValue>();

  /** Emits when the user cancels an edit. */
  readonly cancelled = output<void>();

  readonly isEdit = computed(() => this.task() !== null);

  /** True while an add/update is in flight; disables submit to avoid double-submits. */
  readonly submitting = inject(TaskStore).submitting;

  private readonly formDir = viewChild(FormGroupDirective);

  private readonly fb = inject(FormBuilder);
  readonly form = this.fb.nonNullable.group({
    title: ['', [Validators.required, Validators.maxLength(200)]],
    done: [false],
  });

  constructor() {
    // Keep the form in sync with the `task` input (side effect, not derivation).
    effect(() => {
      const t = this.task();
      if (t) {
        this.form.setValue({ title: t.title, done: t.done });
      } else {
        this.form.reset({ title: '', done: false });
      }
    });
  }

  submit(): void {
    if (this.form.invalid) {
      // Don't block submission by disabling the button — surface errors instead.
      this.form.markAllAsTouched();
      return;
    }
    const { title, done } = this.form.getRawValue();
    this.save.emit({ title: title.trim(), done });
    if (!this.isEdit()) {
      // resetForm() (not form.reset()) also clears the submitted state, so the
      // now-empty required field doesn't immediately show as an error.
      this.formDir()?.resetForm({ title: '', done: false });
    }
  }
}
