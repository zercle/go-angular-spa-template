/** A task as returned by the API (mirrors the backend taskResponse DTO). */
export interface Task {
  id: number;
  title: string;
  done: boolean;
  createdAt: string;
  updatedAt: string;
}

/** Payload to create a task. */
export interface CreateTask {
  title: string;
}

/** Payload to update a task. */
export interface UpdateTask {
  title: string;
  done: boolean;
}

/** The API wraps successful responses as { "data": ... }. */
export interface Envelope<T> {
  data: T;
}
