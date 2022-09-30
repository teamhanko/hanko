import { Headers, HttpClient } from "../../../src/lib/client/HttpClient";
import { RequestTimeoutError, TechnicalError } from "../../../src";
import Cookies from "js-cookie";

describe("httpClient._fetch()", () => {
  beforeEach(() => {
    Object.defineProperty(global, "XMLHttpRequest", {
      value: jest.fn().mockImplementation(() => ({
        response: JSON.stringify({ foo: "bar" }),
        open: jest.fn(),
        setRequestHeader: jest.fn(),
        getResponseHeader: jest.fn(),
        send: jest.fn(),
      })),
      configurable: true,
      writable: true,
    });
  });

  it("should perform http requests", async () => {
    const xhr = new XMLHttpRequest();
    const client = new HttpClient("http://test.api");

    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.onload();
    });

    const response = await client._fetch("/test", { method: "GET" }, xhr);

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
    const jwt = "test-token";
    const xhr = new XMLHttpRequest();
    const client = new HttpClient("http://test.api");

    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.onload();
    });

    Cookies.get = jest.fn().mockReturnValue(jwt);

    await client._fetch("/test", { method: "GET" }, xhr);

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
    const jwt = "test-token";
    const xhr = new XMLHttpRequest();
    const client = new HttpClient("https://test.api");

    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.onload();
    });

    jest.spyOn(xhr, "getResponseHeader").mockReturnValue(jwt);

    Cookies.set = jest.fn();

    await client._fetch("/test", { method: "GET" }, xhr);

    expect(xhr.getResponseHeader).toHaveBeenCalledWith("X-Auth-Token");
    expect(Cookies.set).toHaveBeenCalledWith("hanko", jwt, { secure: true });
  });

  it("should handle onerror", async () => {
    const xhr = new XMLHttpRequest();
    const client = new HttpClient("http://test.api");

    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.onerror();
    });

    const response = client._fetch("/test", { method: "GET" }, xhr);

    await expect(response).rejects.toThrow(TechnicalError);
  });

  it("should handle ontimeout", async () => {
    const xhr = new XMLHttpRequest();
    const client = new HttpClient("http://test.api");

    jest.spyOn(xhr, "send").mockImplementation(function () {
      // eslint-disable-next-line no-invalid-this
      this.ontimeout();
    });

    const response = client._fetch("/test", { method: "GET" }, xhr);

    await expect(response).rejects.toThrow(RequestTimeoutError);
  });
});

describe("httpClient.get()", () => {
  it("should call get with correct args", async () => {
    const client = new HttpClient("http://test.api");

    client._fetch = jest.fn();
    await client.get("/test");

    expect(client._fetch).toHaveBeenCalledWith("/test", { method: "GET" });
  });
});

describe("httpClient.post()", () => {
  it("should call post with correct args", async () => {
    const client = new HttpClient("http://test.api");

    client._fetch = jest.fn();
    await client.post("/test");

    expect(client._fetch).toHaveBeenCalledWith("/test", { method: "POST" });
  });
});

describe("httpClient.put()", () => {
  it("should call put with correct args", async () => {
    const client = new HttpClient("http://test.api");

    client._fetch = jest.fn();
    await client.put("/test");

    expect(client._fetch).toHaveBeenCalledWith("/test", { method: "PUT" });
  });
});

describe("headers.get()", () => {
  it("should return headers", async () => {
    const xhr = new XMLHttpRequest();
    const header = new Headers(xhr);

    jest.spyOn(xhr, "getResponseHeader").mockReturnValue("bar");

    expect(header.get("foo")).toEqual("bar");
  });
});
