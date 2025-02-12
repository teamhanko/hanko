import { SessionClient, UserClient, HttpClient, Relay, Hanko } from "../src";

describe("class hanko", () => {
  it("should hold instances of available Hanko API clients", async () => {
    const hanko = new Hanko("http://api.test");

    expect(hanko.session).toBeInstanceOf(SessionClient);
    expect(hanko.client).toBeInstanceOf(HttpClient);
    expect(hanko.relay).toBeInstanceOf(Relay);
    expect(hanko.user).toBeInstanceOf(UserClient);
  });
});
