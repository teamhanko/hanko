import { Throttle } from "../../src/lib/Throttle";

describe("throttle()", () => {
  let mockFn: jest.Mock;

  beforeEach(() => {
    jest.useFakeTimers();
    mockFn = jest.fn();
  });

  afterEach(() => {
    jest.clearAllMocks();
    jest.clearAllTimers();
    jest.useRealTimers();
  });

  it("should throttle the function to once per specified interval", async () => {
    const throttledFn = Throttle.throttle(mockFn, 1000);

    expect(mockFn).not.toBeCalled();

    throttledFn();
    expect(mockFn).toHaveBeenCalledTimes(1);

    throttledFn();
    expect(mockFn).toHaveBeenCalledTimes(1);

    jest.advanceTimersByTime(1000);
    expect(mockFn).toHaveBeenCalledTimes(2);
  });

  it("should allow leading edge calls to be disabled", async () => {
    const throttledFunction = Throttle.throttle(mockFn, 1000, {
      leading: false,
    });

    expect(mockFn).not.toBeCalled();

    throttledFunction();
    expect(mockFn).not.toBeCalled();

    jest.advanceTimersByTime(500);
    expect(mockFn).not.toBeCalled();

    jest.advanceTimersByTime(500);
    expect(mockFn).toHaveBeenCalledTimes(1);

    throttledFunction();
    jest.advanceTimersByTime(1000);
    expect(mockFn).toHaveBeenCalledTimes(2);
  });

  it("should allow trailing edge calls to be disabled", async () => {
    const throttledFunction = Throttle.throttle(mockFn, 1000, {
      trailing: false,
    });

    expect(mockFn).not.toBeCalled();

    throttledFunction();
    expect(mockFn).toHaveBeenCalledTimes(1);

    throttledFunction();
    throttledFunction();
    throttledFunction();

    jest.advanceTimersByTime(500);
    expect(mockFn).toHaveBeenCalledTimes(1);

    jest.advanceTimersByTime(500);
    expect(mockFn).toHaveBeenCalledTimes(1);

    throttledFunction();
    jest.advanceTimersByTime(1000);
    expect(mockFn).toHaveBeenCalledTimes(2);
  });
});
