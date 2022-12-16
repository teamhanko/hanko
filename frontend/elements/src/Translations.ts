export const translations = {
  en: {
    headlines: {
      error: "An error has occurred",
      loginEmail: "Sign in or sign up",
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
      createdAt: "Created at",
      connectedAccounts: "Connected accounts",
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
        "Your email addresses are used for communication and authentication.",
      changePassword: "Set a new password.",
      managePasskeys: "Your passkeys allow you to sign in to this account.",
      isPrimaryEmail:
        "Used for communication, passcodes, and as username for passkeys. To change the primary email address, add another email address first and set it as primary.",
      setPrimaryEmail:
        "Set this email address primary so it will be used for communications, for passcodes, and as a username for passkeys.",
      emailVerified: "This email address has been verified.",
      emailUnverified: "This email address has not been verified.",
      emailDelete:
        "If you delete this email address, it can no longer be used for signing in to your account. Passkeys that may have been created with this email address will remain intact.",
      emailDeleteThirdPartyConnection:
        "If you delete this email address, it can no longer be used for signing in. You can also no longer sign in with or reconnect your {provider} account. Passkeys that may have been created with this email address will remain intact.",
      emailDeletePrimary:
        "The primary email address cannot be deleted. Add another email address first and make it your primary email address.",
      renamePasskey:
        "Set a name for the passkey that helps you identify where it is stored.",
      deletePasskey:
        "Delete this passkey from your account. Note that the passkey will still exist on your devices and needs to be deleted there as well.",
    },
    labels: {
      or: "or",
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
  },
  de: {
    headlines: {
      error: "Ein Fehler ist aufgetreten",
      loginEmail: "Anmelden / Registrieren",
      loginFinished: "Login erfolgreich",
      loginPasscode: "Passcode eingeben",
      loginPassword: "Passwort eingeben",
      registerAuthenticator: "Passkey einrichten",
      registerConfirm: "Konto erstellen?",
      registerPassword: "Neues Passwort eingeben",
      profileEmails: "E-Mails",
      profilePassword: "Passwort",
      profilePasskeys: "Passkeys",
      isPrimaryEmail: "Primäre E-Mail-Adresse",
      setPrimaryEmail: "Als primäre E-Mail-Adresse festlegen",
      emailVerified: "Verifiziert",
      emailUnverified: "Unverifiziert",
      emailDelete: "Löschen",
      renamePasskey: "Passkey umbenennen",
      deletePasskey: "Passkey löschen",
      createdAt: "Erstellt am",
      connectedAccounts: "Verbundene Konten",
    },
    texts: {
      enterPasscode:
        'Geben Sie den Passcode ein, der an die E-Mail-Adresse "{emailAddress}" gesendet wurde.',
      setupPasskey:
        "Ihr Gerät unterstützt die sichere Anmeldung mit Passkeys. Hinweis: Ihre biometrischen Daten verbleiben sicher auf Ihrem Gerät und werden niemals an unseren Server gesendet.",
      createAccount:
        'Es existiert kein Konto für "{emailAddress}". Möchten Sie ein neues Konto erstellen?',
      passwordFormatHint:
        "Das Passwort muss zwischen {minLength} und {maxLength} Zeichen lang sein.",
      manageEmails:
        "Ihre E-Mail-Adressen werden zur Kommunikation und Authentifizierung verwendet.",
      changePassword: "Setze ein neues Passwort.",
      managePasskeys:
        "Passkeys können für die Anmeldung bei diesem Account verwendet werden.",
      isPrimaryEmail:
        "Wird für die Kommunikation, Passcodes und als Benutzername für Passkeys verwendet. Um die primäre E-Mail-Adresse zu ändern, fügen Sie zuerst eine andere E-Mail-Adresse hinzu und legen Sie sie als primär fest.",
      setPrimaryEmail:
        "Legen Sie diese E-Mail-Adresse als primär fest, damit sie für die Kommunikation, für Passcodes und als Benutzername für Passkeys genutzt wird.",
      emailVerified: "Diese E-Mail-Adresse wurde verifiziert.",
      emailUnverified: "Diese E-Mail-Adresse wurde noch nicht verifiziert.",
      emailDelete:
        "Wenn Sie diese E-Mail-Adresse löschen, kann sie nicht mehr für die Anmeldung bei Ihrem Konto verwendet werden. Passkeys, die möglicherweise mit dieser E-Mail-Adresse erstellt wurden, funktionieren weiterhin.",
      emailDeleteThirdPartyConnection:
        "Wenn Sie diese E-Mail-Adresse löschen, kann sie nicht mehr für die Anmeldung bei Ihrem Konto verwendet werden. Sie können das verbundene {provider}-Konto ebenfalls nicht mehr zu Anmeldung nutzen oder dieses neu verbinden. Passkeys, die möglicherweise mit dieser E-Mail-Adresse erstellt wurden, funktionieren weiterhin.",
      emailDeletePrimary:
        "Die primäre E-Mail-Adresse kann nicht gelöscht werden. Fügen Sie zuerst eine andere E-Mail-Adresse hinzu und legen Sie diese als primär fest.",
      renamePasskey:
        "Legen Sie einen Namen für den Passkey fest, anhand dessen Sie erkennen können, wo er gespeichert ist.",
      deletePasskey:
        "Löschen Sie diesen Passkey aus Ihrem Konto. Beachten Sie, dass der Passkey noch auf Ihren Geräten vorhanden ist und auch dort gelöscht werden muss.",
    },
    labels: {
      or: "oder",
      email: "E-Mail",
      continue: "Weiter",
      skip: "Überspringen",
      save: "Speichern",
      password: "Passwort",
      signInPassword: "Mit einem Passwort anmelden",
      signInPasscode: "Mit einem Passcode anmelden",
      forgotYourPassword: "Passwort vergessen?",
      back: "Zurück",
      signInPasskey: "Anmelden mit Passkey",
      registerAuthenticator: "Passkey einrichten",
      signIn: "Anmelden",
      signUp: "Registrieren",
      sendNewPasscode: "Neuen Code senden",
      passwordRetryAfter: "Neuer Versuch in {passwordRetryAfter}",
      passcodeResendAfter: "Neuen Code in {passcodeResendAfter} anfordern",
      unverifiedEmail: "unverifiziert",
      primaryEmail: "primär",
      setAsPrimaryEmail: "Als primär festlegen",
      verify: "Verifizieren",
      delete: "Löschen",
      newEmailAddress: "Neue E-Mail-Adresse",
      newPassword: "Neues Passwort",
      rename: "Umbenennen",
      newPasskeyName: "Neuer Passkey Name",
      addEmail: "E-Mail-Adresse hinzufügen",
      changePassword: "Password ändern",
      addPasskey: "Passkey hinzufügen",
      webauthnUnsupported:
        "Passkeys werden von ihrem Browser nicht unterrstützt",
      signInWith: "Anmelden mit {provider}",
    },
    errors: {
      somethingWentWrong:
        "Ein technischer Fehler ist aufgetreten. Bitte versuchen Sie es später erneut.",
      requestTimeout: "Die Anfrage hat das Zeitlimit überschritten.",
      invalidPassword: "E-Mail-Adresse oder Passwort falsch.",
      invalidPasscode: "Der Passcode war nicht richtig.",
      passcodeAttemptsReached:
        "Der Passcode wurde zu oft falsch eingegeben. Bitte fragen Sie einen neuen Code an.",
      tooManyRequests:
        "Es wurden zu viele Anfragen gestellt. Bitte warten Sie, um den gewünschten Vorgang zu wiederholen.",
      unauthorized:
        "Ihre Sitzung ist abgelaufen. Bitte melden Sie sich erneut an.",
      invalidWebauthnCredential:
        "Dieser Passkey kann nicht mehr verwendet werden.",
      passcodeExpired:
        "Der Passcode ist abgelaufen. Bitte fordern Sie einen neuen Code an.",
      userVerification:
        "Nutzer-Verifikation erforderlich. Bitte stellen Sie sicher, dass Ihr Gerät durch eine PIN oder Biometrie abgesichert ist.",
      emailAddressAlreadyExistsError: "Die E-Mail-Adresse existiert bereits.",
      maxNumOfEmailAddressesReached:
        "Es können keine weiteren E-Mail-Adressen hinzugefügt werden.",
      thirdPartyAccessDenied:
        "Zugriff verweigert. Die Anfrage wurde durch den Nutzer abgebrochen oder der Provider hat den Zugriff aus anderen Gründen verweigert.",
      thirdPartyMultipleAccounts:
        "Konto kann nicht eindeutig identifiziert werden. Die genutzte E-Mail-Adresse wird von mehreren Konten verwendet.",
      thirdPartyUnverifiedEmail:
        "Verifizierung der E-Mail-Adresse erforderlich. Bitte verifizieren sie die genutzte E-Mail-Adresse bei ihrem Provider.",
    },
  },
};
