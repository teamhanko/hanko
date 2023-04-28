/**
 * Represents a task to be scheduled in the scheduler.
 *
 * @ignore
 * @param {number} timeoutID - the ID returned by `setTimeout()`
 * @param {string} type - The type, helps to organize the tasks.
 * @param {Function} func - The function to be called within the task.
 */
export type Task = {
  timeoutID: number;
  type: string;
  func: () => any;
};

/**
 * A class that manages scheduled tasks.
 *
 * @category SDK
 * @subcategory Internal
 */
export class Scheduler {
  // An array of scheduled tasks.
  _tasks: Task[] = [];

  /**
   * Removes a task from the scheduler.
   *
   * @private
   * @param {Task} task - The task to be removed.
   */
  private removeTask(task: Task) {
    window.clearTimeout(task.timeoutID);
    this._tasks = this._tasks.filter((_task) => _task !== task);
  }

  /**
   * Removes all tasks with a given type from the scheduler.
   *
   * @param {string} type - The type of tasks to be removed.
   */
  removeTasksWithType(type: string) {
    const tasks = this._tasks.filter((task) => task.type === type);
    tasks.forEach((task) => this.removeTask(task));
  }

  /**
   * Schedules a task to be executed after a given timeout.
   *
   * @param {string} type - The type of the task.
   * @param {Function} func - The function to be executed when the task is triggered.
   * @param {number} timeoutSeconds - The timeout after which the task should be executed, in seconds.
   */
  scheduleTask(type: string, func: () => any, timeoutSeconds: number) {
    const task: Task = {
      timeoutID: window.setTimeout(() => {
        func();
        this.removeTask(task);
      }, timeoutSeconds * 1000),
      type,
      func,
    };
    this._tasks.push(task);
  }
}
