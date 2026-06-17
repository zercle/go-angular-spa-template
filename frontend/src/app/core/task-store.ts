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

  /** Read-only views exposed to components. */
  readonly tasks = this._tasks.asReadonly();
  readonly loading = this._loading.asReadonly();
  readonly error = this._error.asReadonly();

  /** Derived count of not-yet-done tasks. */
  readonly remaining = computed(() => this._tasks().filter((t) => !t.done).length);

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

  /** Creates a task and prepends it to state. */
  async add(input: CreateTask): Promise<void> {
    const created = await firstValueFrom(this.http.post<Task>(this.baseUrl, input));
    this._tasks.update((list) => [created, ...list]);
  }

  /** Updates a task and replaces it in state. */
  async update(id: string, input: UpdateTask): Promise<void> {
    const updated = await firstValueFrom(this.http.put<Task>(`${this.baseUrl}/${id}`, input));
    this._tasks.update((list) => list.map((t) => (t.id === id ? updated : t)));
  }

  /** Toggles the done flag of a task. */
  async toggle(task: Task): Promise<void> {
    await this.update(task.id, { title: task.title, done: !task.done });
  }

  /** Deletes a task and removes it from state. */
  async remove(id: string): Promise<void> {
    await firstValueFrom(this.http.delete<void>(`${this.baseUrl}/${id}`));
    this._tasks.update((list) => list.filter((t) => t.id !== id));
  }
}
