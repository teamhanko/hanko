import { Translation } from "./translations";

export const en: Translation = {
  headlines: {
    error: "An error has occurred",
    loginEmail: "Sign in or sign up",
    loginEmailNoSignup: "Sign in",
    loginFinished: "Login successful",
    loginPasscode: "Enter passcode",
    loginPassword: "Enter password",
    registerAuthenticator: "Save a passkey",
    registerConfirm: "Create account?",
    registerPassword: "Set new password",
    profileEmails: "Emails",
    profilePassword: "Password",
    profilePasskeys: "Passkeys",
    isPrimaryEmail: "Primary email address",
    setPrimaryEmail: "Set primary email address",
    emailVerified: "Verified",
    emailUnverified: "Unverified",
    emailDelete: "Delete",
    renamePasskey: "Rename passkey",
    deletePasskey: "Delete passkey",
    lastUsedAt: "Last used at",
    createdAt: "Created at",
    connectedAccounts: "Connected accounts",
    deleteAccount: "Delete account",
    accountNotFound: "Account not found",
  },
  texts: {
    enterPasscode: 'Enter the passcode that was sent to "{emailAddress}".',
    setupPasskey:
      "Sign in to your account easily and securely with a passkey. Note: Your biometric data is only stored on your devices and will never be shared with anyone.",
    createAccount:
      'No account exists for "{emailAddress}". Do you want to create a new account?',
    passwordFormatHint:
      "Must be between {minLength} and {maxLength} characters long.",
    manageEmails:
      "Used for passcode authentication.",
    changePassword: "Set a new password.",
    managePasskeys: "Your passkeys allow you to sign in to this account.",
    isPrimaryEmail:
      "This email address will be used as username for your passkeys.",
    setPrimaryEmail:
      "Set this email to be used as username for new passkeys.",
    emailVerified: "This email address has been verified.",
    emailUnverified: "This email address has not been verified.",
    emailDelete:
      "If you delete this email address, it can no longer be used to sign in.",
    emailDeleteThirdPartyConnection:
      "If you delete this email address, it can no longer be used to sign in.",
    emailDeletePrimary:
      "The primary email address cannot be deleted.",
    renamePasskey:
      "Set a name for the passkey.",
    deletePasskey:
      "Delete this passkey from your account.",
    deleteAccount:
      "Are you sure you want to delete this account? All data will be deleted immediately and cannot be recovered.",
    noAccountExists:
      'No account exists for "{emailAddress}".',
  },
  labels: {
    or: "or",
    no: "no",
    yes: "yes",
    email: "Email",
    continue: "Continue",
    skip: "Skip",
    save: "Save",
    password: "Password",
    signInPassword: "Sign in with a password",
    signInPasscode: "Sign in with a passcode",
    forgotYourPassword: "Forgot your password?",
    back: "Back",
    signInPasskey: "Sign in with a passkey",
    registerAuthenticator: "Save a passkey",
    signIn: "Sign in",
    signUp: "Sign up",
    sendNewPasscode: "Send new code",
    passwordRetryAfter: "Retry in {passwordRetryAfter}",
    passcodeResendAfter: "Request a new code in {passcodeResendAfter}",
    unverifiedEmail: "unverified",
    primaryEmail: "primary",
    setAsPrimaryEmail: "Set as primary",
    verify: "Verify",
    delete: "Delete",
    newEmailAddress: "New email address",
    newPassword: "New password",
    rename: "Rename",
    newPasskeyName: "New passkey name",
    addEmail: "Add email",
    changePassword: "Change password",
    addPasskey: "Add passkey",
    webauthnUnsupported: "Passkeys are not supported by your browser",
    signInWith: "Sign in with {provider}",
    deleteAccount: "Yes, delete this account.",
  },
  errors: {
    somethingWentWrong:
      "A technical error has occurred. Please try again later.",
    requestTimeout: "The request timed out.",
    invalidPassword: "Wrong email or password.",
    invalidPasscode: "The passcode provided was not correct.",
    passcodeAttemptsReached:
      "The passcode was entered incorrectly too many times. Please request a new code.",
    tooManyRequests:
      "Too many requests have been made. Please wait to repeat the requested operation.",
    unauthorized: "Your session has expired. Please log in again.",
    invalidWebauthnCredential: "This passkey cannot be used anymore.",
    passcodeExpired: "The passcode has expired. Please request a new one.",
    userVerification:
      "User verification required. Please ensure your authenticator device is protected with a PIN or biometric.",
    emailAddressAlreadyExistsError: "The email address already exists.",
    maxNumOfEmailAddressesReached: "No further email addresses can be added.",
    thirdPartyAccessDenied:
      "Access denied. The request was cancelled by the user or the provider has denied access for other reasons.",
    thirdPartyMultipleAccounts:
      "Cannot identify account. The email address is used by multiple accounts.",
    thirdPartyUnverifiedEmail:
      "Email verification required. Please verify the used email address with your provider.",
  },
};
