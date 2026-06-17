import { Service, computed, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom, map } from 'rxjs';

import { CreateTask, ListTasksResponse, Task, UpdateTask } from './task';

/**
 * Signal-based store for the tasks resource. Holds the canonical client-side
 * state (tasks, loading, error) as signals and talks to the versioned API.
 */
@Service()
export class TaskStore {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = '/api/v1/tasks';

  private readonly _tasks = signal<Task[]>([]);
  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);
  private readonly _submitting = signal(false);
  private readonly _pendingIds = signal<ReadonlySet<string>>(new Set());

  /** Read-only views exposed to components. */
  readonly tasks = this._tasks.asReadonly();
  readonly loading = this._loading.asReadonly();
  readonly error = this._error.asReadonly();

  /** True while an add/update is in flight (guards the form submit button). */
  readonly submitting = this._submitting.asReadonly();

  /** Derived count of not-yet-done tasks. */
  readonly remaining = computed(() => this._tasks().filter((t) => !t.done).length);

  /** Whether a per-task mutation (toggle/delete) is in flight for `id`. */
  pending(id: string): boolean {
    return this._pendingIds().has(id);
  }

  /** Fetches all tasks into state (list endpoint returns { tasks: [...] }). */
  async load(): Promise<void> {
    this._loading.set(true);
    this._error.set(null);
    try {
      const tasks = await firstValueFrom(
        this.http.get<ListTasksResponse>(this.baseUrl).pipe(map((r) => r.tasks)),
      );
      this._tasks.set(tasks);
    } catch {
      this._error.set('Failed to load tasks.');
    } finally {
      this._loading.set(false);
    }
  }

  /** Creates a task and prepends it to state. Returns true on success. */
  async add(input: CreateTask): Promise<boolean> {
    if (this._submitting()) return false;
    this._submitting.set(true);
    try {
      const created = await firstValueFrom(this.http.post<Task>(this.baseUrl, input));
      this._tasks.update((list) => [created, ...list]);
      this._error.set(null);
      return true;
    } catch {
      this._error.set('Failed to add task.');
      return false;
    } finally {
      this._submitting.set(false);
    }
  }

  /** Updates a task and replaces it in state. Returns true on success. */
  async update(id: string, input: UpdateTask): Promise<boolean> {
    if (this._submitting()) return false;
    this._submitting.set(true);
    try {
      const updated = await firstValueFrom(this.http.put<Task>(`${this.baseUrl}/${id}`, input));
      this._tasks.update((list) => list.map((t) => (t.id === id ? updated : t)));
      this._error.set(null);
      return true;
    } catch {
      this._error.set('Failed to update task.');
      return false;
    } finally {
      this._submitting.set(false);
    }
  }

  /** Toggles the done flag of a task. Returns true on success. */
  async toggle(task: Task): Promise<boolean> {
    if (this.pending(task.id)) return false;
    this.markPending(task.id, true);
    try {
      const updated = await firstValueFrom(
        this.http.put<Task>(`${this.baseUrl}/${task.id}`, {
          title: task.title,
          done: !task.done,
        }),
      );
      this._tasks.update((list) => list.map((t) => (t.id === task.id ? updated : t)));
      this._error.set(null);
      return true;
    } catch {
      this._error.set('Failed to update task.');
      return false;
    } finally {
      this.markPending(task.id, false);
    }
  }

  /** Deletes a task and removes it from state. Returns true on success. */
  async remove(id: string): Promise<boolean> {
    if (this.pending(id)) return false;
    this.markPending(id, true);
    try {
      await firstValueFrom(this.http.delete<void>(`${this.baseUrl}/${id}`));
      this._tasks.update((list) => list.filter((t) => t.id !== id));
      this._error.set(null);
      return true;
    } catch {
      this._error.set('Failed to delete task.');
      return false;
    } finally {
      this.markPending(id, false);
    }
  }

  /** Adds or removes `id` from the set of tasks with an in-flight mutation. */
  private markPending(id: string, pending: boolean): void {
    this._pendingIds.update((ids) => {
      const next = new Set(ids);
      if (pending) {
        next.add(id);
      } else {
        next.delete(id);
      }
      return next;
    });
  }
}
