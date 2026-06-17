/** A task as returned by the API (mirrors the backend TaskResponse DTO). */
export interface Task {
  id: string;
  title: string;
  done: boolean;
  created_at: string;
  updated_at: string;
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

/** The list endpoint wraps results as { "tasks": [...] }. */
export interface ListTasksResponse {
  tasks: Task[];
}
