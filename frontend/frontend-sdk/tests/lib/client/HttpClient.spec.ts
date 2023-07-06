import {
  Headers,
  Response,
  HttpClient,
} from "../../../src/lib/client/HttpClient";
import { RequestTimeoutError, TechnicalError } from "../../../src";

const jwt = "test-token";
let httpClient: HttpClient;
let xhr: XMLHttpRequest;

beforeEach(() => {
  Object.defineProperty(global, "XMLHttpRequest", {
    value: jest.fn().mockImplementation(() => ({
      response: JSON.stringify({ foo: "bar" }),
      open: jest.fn(),
      setRequestHeader: jest.fn(),
      getResponseHeader: jest.fn(),
      getAllResponseHeaders: jest.fn().mockReturnValue(`X-Auth-Token: ${jwt}`),
      send: jest.fn(),
    })),
    configurable: true,
    writable: true,
  });

  httpClient = new HttpClient("http://test.api", {
    cookieName: "hanko",
    localStorageKey: "hanko",
    timeout: 13000,
  });
  xhr = new XMLHttpRequest();
});

describe("httpClient._fetch()", () => {
  it("should perform http requests", async () => {
    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.onload();
    });

    const response = await httpClient._fetch("/test", { method: "GET" }, xhr);

    expect(xhr.setRequestHeader).toHaveBeenCalledWith(
      "Accept",
      "application/json"
    );
    expect(xhr.setRequestHeader).toHaveBeenCalledWith(
      "Content-Type",
      "application/json"
    );
    expect(xhr.setRequestHeader).toHaveBeenCalledTimes(2);
    expect(xhr.open).toHaveBeenNthCalledWith(
      1,
      "GET",
      "http://test.api/test",
      true
    );
    expect(response.json()).toEqual({ foo: "bar" });
  });

  it("should set authorization request headers when cookie is available", async () => {
    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.onload();
    });

    jest.spyOn(httpClient.cookie, "getAuthCookie").mockReturnValue(jwt);

    await httpClient._fetch("/test", { method: "GET" }, xhr);

    expect(xhr.setRequestHeader).toHaveBeenCalledWith(
      "Authorization",
      `Bearer ${jwt}`
    );
    expect(xhr.setRequestHeader).toHaveBeenCalledTimes(3);
  });

  it("should handle onerror", async () => {
    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.onerror();
    });

    const response = httpClient._fetch("/test", { method: "GET" }, xhr);

    await expect(response).rejects.toThrow(TechnicalError);
  });

  it("should handle ontimeout", async () => {
    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.ontimeout();
    });

    const response = httpClient._fetch("/test", { method: "GET" }, xhr);

    await expect(response).rejects.toThrow(RequestTimeoutError);
  });
});

describe("httpClient.get()", () => {
  it("should call get with correct args", async () => {
    httpClient._fetch = jest.fn();
    await httpClient.get("/test");

    expect(httpClient._fetch).toHaveBeenCalledWith("/test", { method: "GET" });
  });
});

describe("httpClient.post()", () => {
  it("should call post with correct args", async () => {
    httpClient._fetch = jest.fn();
    await httpClient.post("/test");

    expect(httpClient._fetch).toHaveBeenCalledWith("/test", { method: "POST" });
  });
});

describe("httpClient.put()", () => {
  it("should call put with correct args", async () => {
    httpClient._fetch = jest.fn();
    await httpClient.put("/test");

    expect(httpClient._fetch).toHaveBeenCalledWith("/test", { method: "PUT" });
  });
});

describe("httpClient.patch()", () => {
  it("should call patch with correct args", async () => {
    httpClient._fetch = jest.fn();
    await httpClient.patch("/test");

    expect(httpClient._fetch).toHaveBeenCalledWith("/test", {
      method: "PATCH",
    });
  });
});

describe("httpClient.delete()", () => {
  it("should call delete with correct args", async () => {
    httpClient._fetch = jest.fn();
    await httpClient.delete("/test");

    expect(httpClient._fetch).toHaveBeenCalledWith("/test", {
      method: "DELETE",
    });
  });
});

describe("headers.get()", () => {
  it("should return headers", async () => {
    const header = new Headers(xhr);

    jest.spyOn(xhr, "getResponseHeader").mockReturnValue("bar");

    expect(header.getResponseHeader("foo")).toEqual("bar");
  });
});

describe("httpClient.processResponseHeadersOnLogin()", () => {
  const jwt = "test-jwt";
  const expirationSeconds = 7;
  const userID = "test-user";

  beforeEach(() => {
    Object.defineProperty(global, "XMLHttpRequest", {
      value: jest.fn().mockImplementation(() => ({
        response: JSON.stringify({ foo: "bar" }),
        open: jest.fn(),
        setRequestHeader: jest.fn(),
        getResponseHeader: jest
          .fn()
          .mockImplementation((name: string) =>
            name === "X-Auth-Token"
              ? jwt
              : name === "X-Session-Lifetime"
              ? `${expirationSeconds}`
              : ""
          ),
        getAllResponseHeaders: jest
          .fn()
          .mockReturnValue("X-Auth-Token: ...\r\nX-Session-Lifetime: ..."),
        send: jest.fn(),
      })),
      configurable: true,
      writable: true,
    });
  });

  it("should set a cookie if x-auth-token response header is available", async () => {
    const client = new HttpClient("http://test.api", {
      cookieName: "hanko",
      localStorageKey: "hanko",
      timeout: 13000,
    });
    const xhr = new XMLHttpRequest();
    const response = new Response(xhr);

    jest.spyOn(response.xhr, "getResponseHeader");
    jest.spyOn(client.passcodeState, "read");
    jest.spyOn(client.passcodeState, "reset");
    jest.spyOn(client.passcodeState, "write");
    jest.spyOn(client.sessionState, "read");
    jest.spyOn(client.cookie, "setAuthCookie");
    jest.spyOn(client.sessionState, "setExpirationSeconds");
    jest.spyOn(client.sessionState, "setUserID");
    jest.spyOn(client.sessionState, "write");

    client.processResponseHeadersOnLogin(userID, response);

    expect(response.xhr.getResponseHeader).toBeCalledTimes(2);
    expect(client.passcodeState.read).toBeCalledTimes(1);
    expect(client.passcodeState.reset).toBeCalledTimes(1);
    expect(client.passcodeState.write).toBeCalledTimes(1);

    expect(client.cookie.setAuthCookie).toHaveBeenCalledTimes(1);
    expect(client.sessionState.read).toHaveBeenCalledTimes(1);
    expect(client.sessionState.setExpirationSeconds).toHaveBeenCalledTimes(1);
    expect(client.sessionState.setUserID).toHaveBeenCalledTimes(1);
    expect(client.sessionState.write).toHaveBeenCalledTimes(1);

    expect(client.sessionState.setExpirationSeconds).toHaveBeenCalledWith(
      expirationSeconds
    );
    expect(client.sessionState.setUserID).toHaveBeenCalledWith(userID);

    expect(client.cookie.setAuthCookie).toHaveBeenCalledWith(jwt, false);
    expect(client.cookie.setAuthCookie).toBeCalledTimes(1);
  });

  it("should set a secure cookie if x-auth-token response header is available and https is being used", async () => {
    const client = new HttpClient("https://test.api", {
      cookieName: "hanko",
      localStorageKey: "hanko",
      timeout: 13000,
    });
    const xhr = new XMLHttpRequest();
    const response = new Response(xhr);

    jest.spyOn(client.cookie, "setAuthCookie");
    client.processResponseHeadersOnLogin(userID, response);

    expect(client.cookie.setAuthCookie).toHaveBeenCalledWith(jwt, true);
    expect(client.cookie.setAuthCookie).toBeCalledTimes(1);
  });
});

describe("response.parseRetryAfterHeader()", () => {
  it.each`
    headerValue  | expected
    ${""}        | ${0}
    ${"0"}       | ${0}
    ${"3"}       | ${3}
    ${"-3"}      | ${-3}
    ${"invalid"} | ${0}
  `("should parse retry-after header", async ({ headerValue, expected }) => {
    const response = new Response(xhr);
    jest.spyOn(xhr, "getResponseHeader").mockReturnValue(headerValue);
    const result = response.parseNumericHeader("Retry-After");
    expect(xhr.getResponseHeader).toHaveBeenCalledWith("Retry-After");
    expect(result).toBe(expected);
  });
});
