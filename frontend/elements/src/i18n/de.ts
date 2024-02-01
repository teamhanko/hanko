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
    emailVerified: "Verifiziert",
    emailUnverified: "Unverifiziert",
    emailDelete: "Löschen",
    renamePasskey: "Passkey umbenennen",
    deletePasskey: "Passkey löschen",
    lastUsedAt: "Zuletzt benutzt am",
    createdAt: "Erstellt am",
    connectedAccounts: "Verbundene Konten",
    deleteAccount: "Konto löschen",
    accountNotFound: "Konto nicht gefunden"
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
    emailDeletePrimary:
      "Die primäre E-Mail-Adresse kann nicht gelöscht werden. Fügen Sie zuerst eine andere E-Mail-Adresse hinzu und legen Sie diese als primär fest.",
    renamePasskey:
      "Legen Sie einen Namen für den Passkey fest, anhand dessen Sie erkennen können, wo er gespeichert ist.",
    deletePasskey:
      "Löschen Sie diesen Passkey aus Ihrem Konto. Beachten Sie, dass der Passkey noch auf Ihren Geräten vorhanden ist und auch dort gelöscht werden muss.",
    deleteAccount:
      "Sind Sie sicher, dass Sie Ihr Konto löschen wollen? Alle Daten werden sofort gelöscht und können nicht wiederhergestellt werden.",
    noAccountExists:
      'Es existiert kein Konto für "{emailAddress}".',
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
    changePassword: "Password ändern",
    createPasskey: "Erstellen Sie einen passkey",
    webauthnUnsupported: "Passkeys werden von ihrem Browser nicht unterrstützt",
    signInWith: "Anmelden mit {provider}",
    deleteAccount: "Ja, dieses Konto löschen.",
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
};
