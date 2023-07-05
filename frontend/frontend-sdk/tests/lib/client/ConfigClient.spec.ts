import { ConfigClient, TechnicalError } from "../../../src";
import { Response } from "../../../src/lib/client/HttpClient";

let configClient: ConfigClient;

beforeEach(() => {
  configClient = new ConfigClient("http://test.api", {
    cookieName: "hanko",
    storageKey: "hanko",
    timeout: 13000,
  });
});

describe("configClient.get()", () => {
  it("should call well-known config endpoint and return config", async () => {
    const response = new Response(new XMLHttpRequest());
    response.ok = true;
    response._decodedJSON = { password: { enabled: true } };

    jest.spyOn(configClient.client, "get").mockResolvedValue(response);
    const config = await configClient.get();
    expect(configClient.client.get).toHaveBeenCalledWith("/.well-known/config");
    expect(config).toEqual(response._decodedJSON);
  });

  it("should throw technical error when API response is not ok", async () => {
    const response = new Response(new XMLHttpRequest());
    configClient.client.get = jest.fn().mockResolvedValue(response);

    const config = configClient.get();
    await expect(config).rejects.toThrow(TechnicalError);
  });

  it("should throw error on API communication failure", async () => {
    configClient.client.get = jest
      .fn()
      .mockRejectedValue(new Error("Test error"));

    const config = configClient.get();
    await expect(config).rejects.toThrowError("Test error");
  });
});
