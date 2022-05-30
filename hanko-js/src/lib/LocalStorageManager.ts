import { Credential } from "./HankoClient";

interface PasscodeState {
  id: string;
  ttl: number;
  resendAfter: number;
}

interface PasswordState {
  retryAfter: number;
}

interface UserState {
  webAuthnCredentials: string[];
  passcode: PasscodeState;
  password: PasswordState;
}

interface UserStates {
  [userID: string]: UserState;
}

interface Store {
  userStates?: UserStates;
}

const initialUserState: UserState = {
  webAuthnCredentials: [],
  passcode: { id: "", ttl: 0, resendAfter: 0 },
  password: { retryAfter: 0 },
};

class LocalStorageManager {
  key: string;

  public constructor(key: string) {
    this.key = key;
  }

  read(): Store {
    let store: Store;
    try {
      const data = localStorage.getItem(this.key);
      const decoded = decodeURIComponent(decodeURI(window.atob(data)));

      store = JSON.parse(decoded);
    } catch (_) {
      return { userStates: {} } as Store;
    }
    return store;
  }

  write(store: Store): void {
    const data = JSON.stringify(store);
    const encoded = window.btoa(encodeURI(encodeURIComponent(data)));
    localStorage.setItem(this.key, encoded);
  }

  getUserState(userID: string) {
    const store = this.read();
    const exists = Object.prototype.hasOwnProperty.call(
      store.userStates,
      userID
    );
    return exists ? store.userStates[userID] : initialUserState;
  }

  setUserState(userID: string, state: UserState) {
    const store = this.read();
    store.userStates[userID] = state;
    this.write(store);
  }

  timeToRemainingSeconds(time: number = 0) {
    return time - Math.floor(Date.now() / 1000);
  }

  remainingSecondsToTime(seconds: number = 0) {
    return Math.floor(Date.now() / 1000) + seconds;
  }
}

export class WebAuthnManager extends LocalStorageManager {
  setCredentialID(userID: string, credentialID: string): void {
    const state = super.getUserState(userID);
    state.webAuthnCredentials.push(credentialID);
    this.setUserState(userID, state);
  }

  matchCredentials(userID: string, match: Credential[]): boolean {
    const { webAuthnCredentials } = super.getUserState(userID);
    const matches = webAuthnCredentials.filter((id) =>
      match.find((c) => c.id === id)
    );
    return matches.length > 0;
  }
}

export class PasscodeManager extends LocalStorageManager {
  getActiveID(userID: string): string {
    const { passcode } = this.getUserState(userID);
    return passcode.id;
  }

  setActiveID(userID: string, passcodeID: string) {
    const state = this.getUserState(userID);
    state.passcode.id = passcodeID;
    this.setUserState(userID, state);
  }

  removeActive(userID: string) {
    const state = this.getUserState(userID);
    state.passcode = initialUserState.passcode;
    this.setUserState(userID, state);
  }

  getTTL(userID: string): number {
    const state = this.getUserState(userID);
    const ttl = this.timeToRemainingSeconds(state.passcode.ttl);
    return ttl > 0 ? ttl : 0;
  }

  setTTL(userID: string, seconds: number): void {
    const state = this.getUserState(userID);
    state.passcode.ttl = this.remainingSecondsToTime(seconds);
    this.setUserState(userID, state);
  }

  getResendAfter(userID: string): number {
    const { passcode } = this.getUserState(userID);
    const resendAfter = this.timeToRemainingSeconds(passcode.resendAfter);
    return resendAfter > 0 ? resendAfter : 0;
  };

  setResendAfter(userID: string, seconds: number): void {
    const state = this.getUserState(userID);
    state.passcode.resendAfter = this.remainingSecondsToTime(seconds);
    this.setUserState(userID, state);
  };
}

export class PasswordManager extends LocalStorageManager {
  getRetryAfter(userID: string): number {
    const state = this.getUserState(userID);
    const retryAfter = this.timeToRemainingSeconds(state.password.retryAfter);
    return retryAfter > 0 ? retryAfter : 0;
  }

  setRetryAfter(userID: string, seconds: number): void {
    const state = this.getUserState(userID);
    state.password.retryAfter = this.remainingSecondsToTime(seconds);
    this.setUserState(userID, state);
  }
}
