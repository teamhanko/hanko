import { UserState } from "./UserState";
import { Credential } from "../Dto";

/**
 * @interface
 * @category SDK
 * @subcategory Internal
 * @property {string[]?} credentials - A list of known credential IDs on the current browser.
 */
export interface LocalStorageWebauthn {
  credentials?: string[];
}

/**
 * A class that manages WebAuthn credentials via local storage.
 *
 * @extends UserState
 * @category SDK
 * @subcategory Internal
 */
class WebauthnState extends UserState {
  /**
   * Gets the WebAuthn state.
   *
   * @private
   * @param {string} userID - The UUID of the user.
   * @return {LocalStorageWebauthn}
   */
  private getState(userID: string): LocalStorageWebauthn {
    return (super.getUserState(userID).webauthn ||= {});
  }

  /**
   * Reads the current state.
   *
   * @public
   * @return {WebauthnState}
   */
  read(): WebauthnState {
    super.read();

    return this;
  }

  /**
   * Gets the list of known credentials on the current browser.
   *
   * @param {string} userID - The UUID of the user.
   * @return {string[]}
   */
  getCredentials(userID: string): string[] {
    return (this.getState(userID).credentials ||= []);
  }

  /**
   * Adds the credential to the list of known credentials.
   *
   * @param {string} userID - The UUID of the user.
   * @param {string} credentialID - The WebAuthn credential ID.
   * @return {WebauthnState}
   */
  addCredential(userID: string, credentialID: string): WebauthnState {
    this.getCredentials(userID).push(credentialID);

    return this;
  }

  /**
   * Returns the intersection between the specified list of credentials and the known credentials stored in
   * the local storage.
   *
   * @param {string} userID - The UUID of the user.
   * @param {Credential[]} match - A list of credential IDs to be matched against the local storage.
   * @return {Credential[]}
   */
  matchCredentials(userID: string, match: Credential[]): Credential[] {
    return this.getCredentials(userID)
      .filter((id) => match.find((c) => c.id === id))
      .map((id: string) => ({ id } as Credential));
  }
}

export { WebauthnState };
