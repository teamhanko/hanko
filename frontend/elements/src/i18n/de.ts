import { Translation } from "./translations";

export const de: Translation = {
  headlines: {
    error: "Ein Fehler ist aufgetreten",
    loginEmail: "Anmelden / Registrieren",
    loginEmailNoSignup: "Anmelden",
    loginFinished: "Login erfolgreich",
    loginPasscode: "Passcode eingeben",
    loginPassword: "Passwort eingeben",
    registerAuthenticator: "Erstellen Sie einen Passkey",
    registerConfirm: "Konto erstellen?",
    registerPassword: "Neues Passwort eingeben",
    otpSetUp: "Authenticator-App einrichten",
    profileEmails: "E-Mails",
    profilePassword: "Passwort",
    profilePasskeys: "Passkeys",
    isPrimaryEmail: "Primäre E-Mail-Adresse",
    setPrimaryEmail: "Als primäre E-Mail-Adresse festlegen",
    createEmail: "Geben Sie eine neue E-Mail-Adresse ein",
    createUsername: "Geben Sie einen neuen Benutzernamen ein",
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
    selectLoginMethod: "Wähle die Anmelde-Methode",
    setupLoginMethod: "Anmelde-Methode einrichten",
    lastUsed: "Zuletzt gesehen",
    ipAddress: "IP Adresse",
    revokeSession: "Sitzung beenden",
    profileSessions: "Sitzungen",
    mfaSetUp: "MFA einrichten",
    securityKeySetUp: "Sicherheitsschlüssel hinzufügen",
    securityKeyLogin: "Sicherheitsschlüssel",
    otpLogin: "Authentifizierungscode",
    renameSecurityKey: "Sicherheitsschlüssel umbenennen",
    deleteSecurityKey: "Sicherheitsschlüssel löschen",
    securityKeys: "Sicherheitsschlüssel",
    authenticatorApp: "Authenticator-App",
    authenticatorAppNotSetUp: "Authenticator-App einrichten",
    authenticatorAppAlreadySetUp: "Authenticator-App ist eingerichtet",
    trustDevice: "Diesem Browser vertrauen?",
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
    otpEnterVerificationCode:
      "Geben Sie den einmaligen Passwort (OTP) ein, den Sie von Ihrer Authenticator-App erhalten haben:",
    otpScanQRCode:
      "Scannen Sie den QR-Code mit Ihrer Authenticator-App (z.B. Google Authenticator oder jede andere TOTP-App). Alternativ können Sie den OTP-Geheimschlüssel manuell in die App eingeben.",
    otpSecretKey: "OTP-Geheimschlüssel",
    passwordFormatHint:
      "Das Passwort muss zwischen {minLength} und {maxLength} Zeichen lang sein.",
    setPrimaryEmail: "Setzen Sie diese E-Mail-Adresse als Kontaktadresse.",
    isPrimaryEmail:
      "Diese E-Mail-Adresse wird verwendet, um Sie bei Bedarf zu kontaktieren.",
    emailVerified: "Diese E-Mail-Adresse wurde verifiziert.",
    emailUnverified: "Diese E-Mail-Adresse wurde noch nicht verifiziert.",
    emailDelete:
      "Wenn Sie diese E-Mail-Adresse löschen, kann sie nicht mehr für die Anmeldung bei Ihrem Konto verwendet werden. Passkeys, die möglicherweise mit dieser E-Mail-Adresse erstellt wurden, funktionieren weiterhin.",
    renamePasskey:
      "Legen Sie einen Namen für den Passkey fest, anhand dessen Sie erkennen können, wo er gespeichert ist.",
    deletePasskey:
      "Löschen Sie diesen Passkey aus Ihrem Konto. Beachten Sie, dass der Passkey noch auf Ihren Geräten vorhanden ist und auch dort gelöscht werden muss.",
    deleteAccount:
      "Sind Sie sicher, dass Sie Ihr Konto löschen wollen? Alle Daten werden sofort gelöscht und können nicht wiederhergestellt werden.",
    noAccountExists: 'Es existiert kein Konto für "{emailAddress}".',
    selectLoginMethodForFutureLogins:
      "Wählen Sie eine der folgenden Anmelde-Methoden aus, um sie für zukünftige Anmeldungen zu verwenden.",
    howDoYouWantToLogin: "Wie möchten Sie sich anmelden?",
    mfaSetUp:
      "Schützen Sie Ihr Konto mit Mehrfaktor-Authentifizierung (MFA). MFA fügt Ihrer Anmeldeprozedur einen zusätzlichen Schritt hinzu, um sicherzustellen, dass Ihr Konto geschützt bleibt, selbst wenn Ihr Passwort oder E-Mail-Konto kompromittiert wird.",
    securityKeyLogin:
      "Verbinden oder aktivieren Sie Ihren Sicherheitsschlüssel und klicken Sie dann auf die Schaltfläche unten. Wenn Sie bereit sind, verwenden Sie USB, NFC oder Ihr Mobilgerät. Befolgen Sie die Anweisungen, um den Anmeldevorgang abzuschließen.",
    otpLogin:
      "Öffnen Sie Ihre Authenticator-App, um den einmaligen Passwort (OTP) zu erhalten. Geben Sie den Code im untenstehenden Feld ein, um sich anzumelden.",
    renameSecurityKey:
      "Legen Sie einen Namen für den Sicherheitsschlüssel fest.",
    deleteSecurityKey:
      "Löschen Sie diesen Sicherheitsschlüssel aus Ihrem Konto.",
    authenticatorAppAlreadySetUp:
      "Ihr Konto ist durch eine Authenticator-App geschützt, die zeitbasierte einmalige Passwörter (TOTP) für die Mehrfaktor-Authentifizierung generiert.",
    authenticatorAppNotSetUp:
      "Schützen Sie Ihr Konto mit einer Authenticator-App, die zeitbasierte einmalige Passwörter (TOTP) für die Mehrfaktor-Authentifizierung generiert.",
    securityKeySetUp:
      "Verwenden Sie einen dedizierten Sicherheitsschlüssel über USB, Bluetooth oder NFC oder Ihr Mobiltelefon. Schließen Sie Ihren Sicherheitsschlüssel an oder aktivieren Sie ihn, und klicken Sie dann auf die Schaltfläche unten und folgen Sie den Anweisungen, um die Registrierung abzuschließen.",
    trustDevice:
      "Wenn Sie diesem Browser vertrauen, müssen Sie bei der nächsten Anmeldung weder Ihr OTP (Einmalpasswort) eingeben noch Ihren Sicherheitsschlüssel für die Multi-Faktor-Authentifizierung (MFA) verwenden.",
  },
  labels: {
    or: "oder",
    no: "nein",
    yes: "ja",
    email: "E-Mail",
    continue: "Weiter",
    copied: "kopiert",
    skip: "Überspringen",
    save: "Speichern",
    password: "Passwort",
    passkey: "Passwort",
    passcode: "Passcode",
    signInPassword: "Mit einem Passwort anmelden",
    signInPasscode: "Mit einem Passcode anmelden",
    forgotYourPassword: "Passwort vergessen?",
    back: "Zurück",
    signInPasskey: "Anmelden mit Passkey",
    registerAuthenticator: "Erstellen Sie einen Passkey",
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
    createPasskey: "Erstellen Sie einen Passkey",
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
    revoke: "Beenden",
    currentSession: "Aktuelle Sitzung",
    authenticatorApp: "Authentifizierungs-App",
    securityKey: "Sicherheitsschlüssel",
    securityKeyUse: "Sicherheitsschlüssel verwenden",
    newSecurityKeyName: "Neuer Sicherheitsschlüsselname",
    createSecurityKey: "Sicherheitsschlüssel hinzufügen",
    authenticatorAppManage: "Authentifizierungs-App verwalten",
    authenticatorAppAdd: "Einrichten",
    configured: "konfiguriert",
    useAnotherMethod: "Eine andere Methode verwenden",
    lastUsed: "Zuletzt verwendet",
    trustDevice: "Diesem Browser vertrauen",
    staySignedIn: "Angemeldet bleiben",
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
    handlerNotFoundError:
      "Der aktuelle Schritt in Ihrem Prozess wird von dieser Anwendungsversion nicht unterstützt.",
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
    unknown_email_error: "Die Email Adresse ist unbekannt.",
    username_already_exists: "Der Benutzername ist bereits vergeben.",
    invalid_username_error:
      "Der Benutzername darf nur Buchstaben, Zahlen und Unterstriche enthalten.",
    email_already_exists: "Die E-Mail-Adresse ist bereits vergeben.",
    not_found: "Die angeforderte Ressource wurde nicht gefunden.",
    operation_not_permitted_error: "Der Vorgang ist nicht erlaubt.",
    flow_discontinuity_error:
      "Der Prozess kann aufgrund der Nutzereinstellungen oder der Konfiguration des Anbieters nicht fortgesetzt werden.",
    form_data_invalid_error:
      "Die übermittelten Formulardaten enthalten Fehler.",
    unauthorized:
      "Ihre Sitzung ist abgelaufen. Bitte melden Sie sich erneut an.",
    value_missing_error: "Der Wert fehlt.",
    value_too_long_error: "Der Wert ist zu lang.",
    value_too_short_error: "Der Wert ist zu kurz.",
    webauthn_credential_invalid_mfa_only:
      "Diese Anmeldeinformation kann nur als zweite Sicherheitsfaktor verwendet werden.",
    webauthn_credential_already_exists:
      "Die Anfrage wurde entweder abgebrochen, abgelaufen oder das Gerät ist bereits registriert. Bitte versuchen Sie es erneut oder verwenden Sie ein anderes Gerät.",
    platform_authenticator_required:
      "Ihr Konto ist so konfiguriert, dass es Plattform-Authentifikatoren verwendet, jedoch unterstützt Ihr aktuelles Gerät oder Ihr Browser diese Funktion nicht. Bitte versuchen Sie es mit einem kompatiblen Gerät oder Browser erneut.",
    third_party_access_denied:
      "Zugriff verweigert. Die Anfrage wurde durch den Nutzer abgebrochen oder der Provider hat den Zugriff aus anderen Gründen verweigert.",
  },
};
