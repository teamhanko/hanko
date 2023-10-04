import { en } from "./en";

export interface Translations {
  [lang: string]: Partial<Translation>;
}

export interface Translation {
  headlines: {
    error: string;
    accountNotFound: string;
    loginEmail: string;
    loginEmailNoSignup: string;
    loginFinished: string;
    loginPasscode: string;
    loginPassword: string;
    registerAuthenticator: string;
    registerConfirm: string;
    registerPassword: string;
    profileEmails: string;
    profilePassword: string;
    profilePasskeys: string;
    isPrimaryEmail: string;
    setPrimaryEmail: string;
    emailVerified: string;
    emailUnverified: string;
    emailDelete: string;
    renamePasskey: string;
    deletePasskey: string;
    lastUsedAt: string;
    createdAt: string;
    connectedAccounts: string;
    deleteAccount: string;
  };
  texts: {
    enterPasscode: string;
    setupPasskey: string;
    createAccount: string;
    noAccountExists: string;
    passwordFormatHint: string;
    manageEmails: string;
    changePassword: string;
    managePasskeys: string;
    isPrimaryEmail: string;
    setPrimaryEmail: string;
    emailVerified: string;
    emailUnverified: string;
    emailDelete: string;
    emailDeleteThirdPartyConnection: string;
    emailDeletePrimary: string;
    renamePasskey: string;
    deletePasskey: string;
    deleteAccount: string;
  };
  labels: {
    or: string;
    no: string;
    yes: string;
    email: string;
    continue: string;
    skip: string;
    save: string;
    password: string;
    signInPassword: string;
    signInPasscode: string;
    forgotYourPassword: string;
    back: string;
    signInPasskey: string;
    registerAuthenticator: string;
    signIn: string;
    signUp: string;
    sendNewPasscode: string;
    passwordRetryAfter: string;
    passcodeResendAfter: string;
    unverifiedEmail: string;
    primaryEmail: string;
    setAsPrimaryEmail: string;
    verify: string;
    delete: string;
    newEmailAddress: string;
    newPassword: string;
    rename: string;
    newPasskeyName: string;
    addEmail: string;
    changePassword: string;
    createPasskey: string;
    webauthnUnsupported: string;
    signInWith: string;
    deleteAccount: string;
  };
  errors: {
    somethingWentWrong: string;
    requestTimeout: string;
    invalidPassword: string;
    invalidPasscode: string;
    passcodeAttemptsReached: string;
    tooManyRequests: string;
    unauthorized: string;
    invalidWebauthnCredential: string;
    passcodeExpired: string;
    userVerification: string;
    emailAddressAlreadyExistsError: string;
    maxNumOfEmailAddressesReached: string;
    thirdPartyAccessDenied: string;
    thirdPartyMultipleAccounts: string;
    thirdPartyUnverifiedEmail: string;
  };
}

export const defaultTranslations: Translations = {
  en,
};
