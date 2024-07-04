import { Translation } from "./translations";

export const de: Translation = {
  headlines: {
    error: "Ein Fehler ist aufgetreten",
    loginEmail: "Anmelden / Registrieren",
    loginEmailNoSignup: "Anmelden",
    loginFinished: "Login erfolgreich",
    loginPasscode: "Passcode eingeben",
    loginPassword: "Passwort eingeben",
    registerAuthenticator: "Erstellen Sie einen passkey",
    registerConfirm: "Konto erstellen?",
    registerPassword: "Neues Passwort eingeben",
    profileEmails: "E-Mails",
    profilePassword: "Passwort",
    profilePasskeys: "Passkeys",
    isPrimaryEmail: "Primäre E-Mail-Adresse",
    setPrimaryEmail: "Als primäre E-Mail-Adresse festlegen",
    createEmail: "Neue E-Mail eingeben",
    createUsername: "Neuen Nutzernamen eingeben",
    emailVerified: "Verifiziert",
    emailUnverified: "Unverifiziert",
    emailDelete: "Löschen",
    renamePasskey: "Passkey umbenennen",
    deletePasskey: "Passkey löschen",
    lastUsedAt: "Zuletzt benutzt am",
    createdAt: "Erstellt am",
    connectedAccounts: "Verbundene Konten",
    deleteAccount: "Konto löschen",
    accountNotFound: "Konto nicht gefunden",
    signIn: "Anmelden",
    signUp: "Registrieren",
  },
  texts: {
    enterPasscode:
      'Geben Sie den Passcode ein, der an die E-Mail-Adresse "{emailAddress}" gesendet wurde.',
    enterPasscodeNoEmail:
      "Geben Sie den Passcode ein, der an Ihre primäre E-Mail-Adresse gesendet wurde.",
    setupPasskey:
      "Ihr Gerät unterstützt die sichere Anmeldung mit Passkeys. Hinweis: Ihre biometrischen Daten verbleiben sicher auf Ihrem Gerät und werden niemals an unseren Server gesendet.",
    createAccount:
      'Es existiert kein Konto für "{emailAddress}". Möchten Sie ein neues Konto erstellen?',
    passwordFormatHint:
      "Das Passwort muss zwischen {minLength} und {maxLength} Zeichen lang sein.",
    isPrimaryEmail:
      "Wird für die Kommunikation, Passcodes und als Benutzername für Passkeys verwendet. Um die primäre E-Mail-Adresse zu ändern, fügen Sie zuerst eine andere E-Mail-Adresse hinzu und legen Sie sie als primär fest.",
    setPrimaryEmail:
      "Legen Sie diese E-Mail-Adresse als primär fest, damit sie für die Kommunikation, für Passcodes und als Benutzername für Passkeys genutzt wird.",
    emailVerified: "Diese E-Mail-Adresse wurde verifiziert.",
    emailUnverified: "Diese E-Mail-Adresse wurde noch nicht verifiziert.",
    emailDelete:
      "Wenn Sie diese E-Mail-Adresse löschen, kann sie nicht mehr für die Anmeldung bei Ihrem Konto verwendet werden. Passkeys, die möglicherweise mit dieser E-Mail-Adresse erstellt wurden, funktionieren weiterhin.",
    emailDeletePrimary:
      "Die primäre E-Mail-Adresse kann nicht gelöscht werden. Fügen Sie zuerst eine andere E-Mail-Adresse hinzu und legen Sie diese als primär fest.",
    renamePasskey:
      "Legen Sie einen Namen für den Passkey fest, anhand dessen Sie erkennen können, wo er gespeichert ist.",
    deletePasskey:
      "Löschen Sie diesen Passkey aus Ihrem Konto. Beachten Sie, dass der Passkey noch auf Ihren Geräten vorhanden ist und auch dort gelöscht werden muss.",
    deleteAccount:
      "Sind Sie sicher, dass Sie Ihr Konto löschen wollen? Alle Daten werden sofort gelöscht und können nicht wiederhergestellt werden.",
    noAccountExists: 'Es existiert kein Konto für "{emailAddress}".',
  },
  labels: {
    or: "oder",
    no: "nein",
    yes: "ja",
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
    registerAuthenticator: "Erstellen Sie einen passkey",
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
    createPasskey: "Erstellen Sie einen passkey",
    webauthnUnsupported: "Passkeys werden von ihrem Browser nicht unterstützt",
    signInWith: "Anmelden mit {provider}",
    deleteAccount: "Ja, dieses Konto löschen.",
    emailOrUsername: "E-Mail oder Nutzername",
    username: "Nutzername",
    optional: "optional",
    dontHaveAnAccount: "Sie haben noch kein Konto?",
    alreadyHaveAnAccount: "Haben Sie bereits ein Konto?",
    changeUsername: "Benutzernamen ändern",
    setUsername: "Benutzernamen setzen",
    changePassword: "Passwort ändern",
    setPassword: "Passwort setzen",
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
    signupDisabled: "Die Kontoregistrierung ist deaktiviert.",
  },
  flowErrors: {
    technical_error:
      "Ein technischer Fehler ist aufgetreten. Bitte versuchen Sie es später erneut.",
    flow_expired_error:
      "Die Sitzung ist abgelaufen, bitte klicken Sie auf die Schaltfläche, um neu zu starten.",
    value_invalid_error: "Der eingegebene Wert ist ungültig.",
    passcode_invalid: "Der angegebene Passcode war nicht korrekt.",
    passkey_invalid: "Dieser Passkey kann nicht mehr verwendet werden.",
    passcode_max_attempts_reached:
      "Der Passcode wurde zu oft falsch eingegeben. Bitte fordern Sie einen neuen Code an.",
    rate_limit_exceeded:
      "Zu viele Anfragen wurden gestellt. Bitte warten Sie, um die angeforderte Operation zu wiederholen.",
    unknown_username_error: "Der Benutzername ist unbekannt.",
    username_already_exists: "Der Benutzername ist bereits vergeben.",
    email_already_taken: "Die E-Mail-Adresse ist bereits vergeben.",
    not_found: "Die angeforderte Ressource wurde nicht gefunden.",
    operation_not_permitted_error: "Der Vorgang ist nicht erlaubt.",
    flow_discontinuity_error:
      "Der Prozess kann aufgrund der Nutzereinstellungen oder der Konfiguration des Anbieters nicht fortgesetzt werden.",
    form_data_invalid_error:
      "Die übermittelten Formulardaten enthalten Fehler.",
  },
};
