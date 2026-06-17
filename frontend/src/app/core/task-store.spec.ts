import { TestBed } from '@angular/core/testing';
import { provideHttpClient, withFetch } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { TaskStore } from './task-store';
import { Task } from './task';

function task(partial: Partial<Task>): Task {
  return { id: 1, title: 't', done: false, createdAt: '', updatedAt: '', ...partial };
}

describe('TaskStore', () => {
  let store: TaskStore;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(withFetch()), provideHttpClientTesting()],
    });
    store = TestBed.inject(TaskStore);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it('loads tasks into state and computes remaining', async () => {
    const promise = store.load();
    const req = httpMock.expectOne('/api/v1/tasks');
    expect(req.request.method).toBe('GET');
    req.flush({ data: [task({ id: 1, done: false }), task({ id: 2, done: true })] });
    await promise;

    expect(store.tasks().length).toBe(2);
    expect(store.remaining()).toBe(1);
  });

  it('prepends a newly created task', async () => {
    const promise = store.add({ title: 'new' });
    const req = httpMock.expectOne('/api/v1/tasks');
    expect(req.request.method).toBe('POST');
    req.flush({ data: task({ id: 99, title: 'new' }) });
    await promise;

    expect(store.tasks()[0].id).toBe(99);
  });

  it('removes a task from state', async () => {
    const loaded = store.load();
    httpMock.expectOne('/api/v1/tasks').flush({ data: [task({ id: 5 })] });
    await loaded;

    const promise = store.remove(5);
    httpMock.expectOne('/api/v1/tasks/5').flush(null);
    await promise;

    expect(store.tasks().length).toBe(0);
  });
});
