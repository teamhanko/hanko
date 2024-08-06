import { EnterpriseClient, Hanko, UserClient } from "../src";

describe("class hanko", () => {
  it("should hold instances of available Hanko API clients", async () => {
    const hanko = new Hanko("http://api.test");

    expect(hanko.user).toBeInstanceOf(UserClient);
    expect(hanko.enterprise).toBeInstanceOf(EnterpriseClient);
  });
});
