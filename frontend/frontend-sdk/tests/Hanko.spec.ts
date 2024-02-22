import {
  ConfigClient,
  EnterpriseClient,
  Hanko,
  PasscodeClient,
  PasswordClient,
  UserClient,
  WebauthnClient,
} from "../src";

describe("class hanko", () => {
  it("should hold instances of available Hanko API clients", async () => {
    const hanko = new Hanko("http://api.test");

    expect(hanko.config).toBeInstanceOf(ConfigClient);
    expect(hanko.user).toBeInstanceOf(UserClient);
    expect(hanko.passcode).toBeInstanceOf(PasscodeClient);
    expect(hanko.password).toBeInstanceOf(PasswordClient);
    expect(hanko.webauthn).toBeInstanceOf(WebauthnClient);
    expect(hanko.enterprise).toBeInstanceOf(EnterpriseClient);
  });
});
