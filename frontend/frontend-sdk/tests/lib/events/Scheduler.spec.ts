import { Scheduler, Task } from "../../../src/lib/events/Scheduler";

describe("Scheduler()", () => {
  let scheduler: Scheduler;

  beforeEach(() => {
    scheduler = new Scheduler();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe("removeTasksWithType", () => {
    it("should remove all tasks with the specified type from the tasks array and clear their timeouts", () => {
      const task1: Task = {
        timeoutID: 123,
        type: "test",
        func: jest.fn(),
      };
      const task2: Task = {
        timeoutID: 456,
        type: "other",
        func: jest.fn(),
      };
      const task3: Task = {
        timeoutID: 789,
        type: "test",
        func: jest.fn(),
      };
      scheduler._tasks = [task1, task2, task3];

      jest.spyOn(global, "clearTimeout");
      scheduler.removeTasksWithType("test");

      expect(scheduler._tasks).toEqual([task2]);
      expect(clearTimeout).toHaveBeenCalledWith(123);
      expect(clearTimeout).toHaveBeenCalledWith(789);
      expect(clearTimeout).not.toHaveBeenCalledWith(456);
    });
  });

  describe("scheduleTask", () => {
    it("should add a new task to the tasks array and schedule its execution", () => {
      const fn = jest.fn();
      jest.spyOn(global, "setTimeout");
      scheduler.scheduleTask("test", fn, 1);

      expect(scheduler._tasks.length).toBe(1);
      expect(scheduler._tasks[0].type).toBe("test");
      expect(scheduler._tasks[0].func).toBe(fn);

      expect(setTimeout).toHaveBeenCalledWith(expect.any(Function), 1000);

      jest.runOnlyPendingTimers();

      expect(fn).toHaveBeenCalled();
      expect(scheduler._tasks.length).toBe(0);
    });
  });
});
