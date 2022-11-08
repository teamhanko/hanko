import { WebauthnSupport } from "../../src/lib/WebauthnSupport";
import { fakePublicKeyCredential } from "../setup";

describe("WebauthnSupport.supported()", () => {
  it("should support webauthn", async () => {
    expect(WebauthnSupport.supported()).toEqual(true);
  });
});

describe("WebauthnSupport.isPlatformAuthenticatorAvailable()", () => {
  it("should pass if platform authenticators are supported", async () => {
    jest
      .spyOn(
        PublicKeyCredential,
        "isUserVerifyingPlatformAuthenticatorAvailable"
      )
      .mockResolvedValueOnce(true);

    const supported = await WebauthnSupport.isPlatformAuthenticatorAvailable();

    expect(supported).toBe(true);
  });

  it("should fail if platform authenticators are unavailable", async () => {
    jest
      .spyOn(
        PublicKeyCredential,
        "isUserVerifyingPlatformAuthenticatorAvailable"
      )
      .mockResolvedValueOnce(false);

    const supported = await WebauthnSupport.isPlatformAuthenticatorAvailable();

    expect(supported).toBe(false);
  });

  it("should fail if webauthn is not supported", async () => {
    jest.spyOn(WebauthnSupport, "supported").mockReturnValueOnce(false);
    const supported = await WebauthnSupport.isPlatformAuthenticatorAvailable();

    expect(supported).toBe(false);
  });
});

describe("WebauthnSupport.isSecurityKeySupported()", () => {
  beforeEach(() => {
    Object.defineProperty(window, "PublicKeyCredential", {
      value: fakePublicKeyCredential,
      configurable: true,
      writable: true,
    });
  });

  it("should pass if security keys are supported", async () => {
    jest
      .spyOn(window.PublicKeyCredential, "isExternalCTAP2SecurityKeySupported")
      .mockResolvedValueOnce(true);

    const supported = await WebauthnSupport.isSecurityKeySupported();

    expect(supported).toBe(true);
  });

  it("should pass if security keys are not supported", async () => {
    jest
      .spyOn(window.PublicKeyCredential, "isExternalCTAP2SecurityKeySupported")
      .mockResolvedValueOnce(false);

    const supported = await WebauthnSupport.isSecurityKeySupported();

    expect(supported).toBe(false);
  });

  it("should fail if webauthn is not supported", async () => {
    window.PublicKeyCredential = undefined;
    jest.spyOn(WebauthnSupport, "supported").mockImplementation(() => false);

    const supported = await WebauthnSupport.isSecurityKeySupported();

    expect(supported).toEqual(false);
    expect(WebauthnSupport.supported).toHaveBeenCalled();
  });
});

describe("WebauthnSupport.isConditionalMediationAvailable()", () => {
  beforeEach(() => {
    Object.defineProperty(window, "PublicKeyCredential", {
      value: fakePublicKeyCredential,
      configurable: true,
      writable: true,
    });
  });

  it("should pass if autofilled requests are supported", async () => {
    jest
      .spyOn(window.PublicKeyCredential, "isConditionalMediationAvailable")
      .mockResolvedValueOnce(true);

    const supported = await WebauthnSupport.isConditionalMediationAvailable();

    expect(supported).toBe(true);
  });

  it("should pass if autofilled requests  are not supported", async () => {
    jest
      .spyOn(window.PublicKeyCredential, "isConditionalMediationAvailable")
      .mockResolvedValueOnce(false);

    const supported = await WebauthnSupport.isConditionalMediationAvailable();

    expect(supported).toBe(false);
  });

  it("should fail if webauthn is not supported", async () => {
    window.PublicKeyCredential = undefined;
    const supported = await WebauthnSupport.isConditionalMediationAvailable();

    expect(supported).toEqual(false);
  });
});
