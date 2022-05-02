import { Credential } from "./HankoClient";

interface UserSettingsData {
  activePasscode?: string;
  webAuthnCredentials?: string[];
  passcodeExpiry?: number;
  passcodeRetryAfter?: number;
  passwordRetryAfter?: number;
}

interface UserSettings {
  [userID: string]: UserSettingsData;
}

interface Store {
  userSettings?: UserSettings;
}

export class LocalStorage {
  key: string;

  public constructor(key: string) {
    this.key = key;
  }

  private read = (): Store => {
    let store: Store;
    try {
      const data = localStorage.getItem(this.key);
      const decoded = decodeURIComponent(decodeURI(window.atob(data)));
      store = JSON.parse(decoded);
    } catch (_) {
      return { userSettings: {} as UserSettingsData } as Store;
    }
    return store;
  };

  private write = (store: Store): void => {
    const data = JSON.stringify(store);
    const encoded = window.btoa(encodeURI(encodeURIComponent(data)));
    localStorage.setItem(this.key, encoded);
  };

  private getUserSettings = (userID: string) => {
    const store = this.read();
    const exists = Object.prototype.hasOwnProperty.call(
      store.userSettings,
      userID
    );
    return exists ? store.userSettings[userID] : {};
  };

  private setUserSettings = (userID: string, settings: UserSettingsData) => {
    const store = this.read();
    store.userSettings[userID] = settings;
    this.write(store);
  };

  public matchCredentials = (
    userID: string,
    match: Credential[]
  ): boolean => {
    const settings = this.getUserSettings(userID);
    const credentials = (settings.webAuthnCredentials ||= []);
    return (
      credentials.filter((id) => match.find((c) => c.id === id)).length > 0
    );
  };

  public getActivePasscodeID = (userID: string): string => {
    const settings = this.getUserSettings(userID);
    return settings.activePasscode;
  };

  public setActivePasscodeID = (userID: string, passcodeID: string) => {
    const settings = this.getUserSettings(userID);
    settings.activePasscode = passcodeID;
    this.setUserSettings(userID, settings);
  };

  public removeActivePasscodeID = (userID: string) => {
    const settings = this.getUserSettings(userID);
    delete settings.activePasscode;
    delete settings.passcodeRetryAfter;
    delete settings.passcodeExpiry;
    this.setUserSettings(userID, settings);
  };

  public setWebAuthnCredentialID = (
    userID: string,
    credentialID: string
  ): void => {
    const settings = this.getUserSettings(userID);
    const credentials = (settings.webAuthnCredentials ||= []);
    credentials.push(credentialID);
    this.setUserSettings(userID, settings);
  };

  public getPasscodeRetryAfter = (userID: string): number => {
    const settings = this.getUserSettings(userID);
    const time = settings.passcodeRetryAfter || 0;
    const retryAfter = time - Math.floor(Date.now() / 1000);
    return retryAfter > 0 ? retryAfter : 0;
  };

  public setPasscodeRetryAfter = (userID: string, seconds: number): void => {
    const settings = this.getUserSettings(userID);
    settings.passcodeRetryAfter = Math.floor(Date.now() / 1000) + seconds;
    this.setUserSettings(userID, settings);
  };

  public getPasswordRetryAfter = (userID: string): number => {
    const settings = this.getUserSettings(userID);
    const time = settings.passwordRetryAfter || 0;
    const retryAfter = time - Math.floor(Date.now() / 1000);
    return retryAfter > 0 ? retryAfter : 0;
  };

  public setPasswordRetryAfter = (userID: string, seconds: number): void => {
    const settings = this.getUserSettings(userID);
    settings.passwordRetryAfter = Math.floor(Date.now() / 1000) + seconds;
    this.setUserSettings(userID, settings);
  };

  public getPasscodeExpiry = (userID: string): number => {
    const settings = this.getUserSettings(userID);
    const time = settings.passcodeExpiry || 0;
    const expires = time - Math.floor(Date.now() / 1000);
    return expires > 0 ? expires : 0;
  };

  public setPasscodeExpiry = (userID: string, seconds: number): void => {
    const settings = this.getUserSettings(userID);
    settings.passcodeExpiry = Math.floor(Date.now() / 1000) + seconds;
    this.setUserSettings(userID, settings);
  };
}
