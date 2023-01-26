import {
  Headers,
  Response,
  HttpClient,
} from "../../../src/lib/client/HttpClient";
import { RequestTimeoutError, TechnicalError } from "../../../src";
import Cookies from "js-cookie";

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

  httpClient = new HttpClient("http://test.api");
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
    expect(xhr.getAllResponseHeaders).toHaveBeenCalledTimes(1);
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

    Cookies.get = jest.fn().mockReturnValue(jwt);

    await httpClient._fetch("/test", { method: "GET" }, xhr);

    expect(xhr.setRequestHeader).toHaveBeenCalledWith(
      "Authorization",
      `Bearer ${jwt}`
    );
    expect(xhr.setRequestHeader).toHaveBeenCalledTimes(3);
  });

  it("should set a cookie if x-auth-token response header is available", async () => {
    const jwt = "test-token";
    const xhr = new XMLHttpRequest();
    const client = new HttpClient("http://test.api");

    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.onload();
    });

    jest.spyOn(xhr, "getResponseHeader").mockReturnValue(jwt);

    Cookies.set = jest.fn();

    await client._fetch("/test", { method: "GET" }, xhr);

    expect(xhr.getResponseHeader).toHaveBeenCalledWith("X-Auth-Token");
    expect(Cookies.set).toHaveBeenCalledWith("hanko", jwt, { secure: false });
  });

  it("should set a secure cookie if x-auth-token response header is available and https is used", async () => {
    httpClient = new HttpClient("https://test.api");

    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.onload();
    });

    jest.spyOn(xhr, "getResponseHeader").mockReturnValue(jwt);

    Cookies.set = jest.fn();

    await httpClient._fetch("/test", { method: "GET" }, xhr);

    expect(xhr.getResponseHeader).toHaveBeenCalledWith("X-Auth-Token");
    expect(Cookies.set).toHaveBeenCalledWith("hanko", jwt, { secure: true });
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

    expect(header.get("foo")).toEqual("bar");
  });
});

describe("response.parseRetryAfterHeader()", () => {
  it.each`
    headerValue  | expected
    ${""}        | ${0}
    ${"0"}       | ${0}
    ${"3"}       | ${3}
    ${"invalid"} | ${0}
  `("should parse retry-after header", async ({ headerValue, expected }) => {
    const response = new Response(xhr);
    jest.spyOn(xhr, "getResponseHeader").mockReturnValue(headerValue);
    const result = response.parseRetryAfterHeader();
    expect(xhr.getResponseHeader).toHaveBeenCalledWith("Retry-After");
    expect(result).toBe(expected);
  });
});
