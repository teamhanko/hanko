export {};

declare global {
  namespace PlaywrightTest {
    interface Matchers<R> {
      toHaveCookie(name?: string): R;
      toHaveLocalStorageEntry(origin?: string, name?: string): R;
      toHaveLocalStorageEntryForUserWithCredential(
        userId: string,
        credentialId: string,
        origin?: string,
        name?: string
      ): R;
      toHaveLocalStorageEntryForUserWithPasscode(
        userId: string,
        passcodeId: string,
        origin?: string,
        name?: string
      ): R;
    }
  }
}
