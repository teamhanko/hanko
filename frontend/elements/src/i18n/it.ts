import { Translation } from "./translations";

export const en: Translation = {
  headlines: {
    error: "Si è verificato un errore An error has occurred",
    loginEmail: "Accedi o registrati",
    loginEmailNoSignup: "Accedi",
    loginFinished: "Accesso riuscito",
    loginPasscode: "Inserisci codice di accesso",
    loginPassword: "Inserisci password",
    registerAuthenticator: "Crea una passkey",
    registerConfirm: "Vuoi creare un account?",
    registerPassword: "Imposta una nuova password",
    profileEmails: "Emails",
    profilePassword: "Password",
    profilePasskeys: "Passkeys",
    isPrimaryEmail: "Indirizzo email principale",
    setPrimaryEmail: "Imposta indirizzo email principale",
    emailVerified: "Verificata",
    emailUnverified: "Non verificata",
    emailDelete: "Rimuovi",
    renamePasskey: "Rinomina passkey",
    deletePasskey: "Elimina passkey",
    lastUsedAt: "Ultimo accesso il",
    createdAt: "Creato il",
    connectedAccounts: "Account collegati",
    deleteAccount: "Rimuovi account",
    accountNotFound: "Account non trovato",
  },
  texts: {
    enterPasscode: 'Inserisci il codice di accesso inviato a "{emailAddress}".',
    setupPasskey:
      "Accedi al tuo accuont in modo semplice e sicuro con una passkey. Nota: I tuoi dati biometrici sono archiviati solo sui tuoi dispositivi e non saranno condivisi con nessuno.",
    createAccount:
      'Nessun account trovato per "{emailAddress}". Vuoi creare un nuovo account?',
    passwordFormatHint:
      "La lunghezza deve essere compresa tra {minLength} e {maxLength} caratteri.",
    manageEmails:
      "Used for passcode authentication.",
    changePassword: "Imposta una nuova password",
    managePasskeys: "Le tue passkey ti consentono di accedere a questo account.",
    isPrimaryEmail:
      "Questo indirizzo email verrà utilizzato come nome utente per le tue passkeys.",
    setPrimaryEmail:
      "Imposa questo indirizzo email per essere usato come nome utente per le tue nuove passkeys.",
    emailVerified: "Questo indirizzo email è stato verificato.",
    emailUnverified: "Questo indirizzo email non è stato verificato.",
    emailDelete:
      "Se cancelli questo indirizzo email, non potrà più essere utilizzato per accedere.",
    emailDeleteThirdPartyConnection:
      "Se cancelli questo indirizzo email, non potrà più essere utilizzato per accedere.",
    emailDeletePrimary:
      "L'indirizzo email principale non può essere cancellato.",
    renamePasskey:
      "Imposta un nome per la passkey.",
    deletePasskey:
      "Cancella questa passkey dal tuo account.",
    deleteAccount:
      "Sei sicuro di voler eliminare questo account? Tutti i dati verranno immediatamente cancellati permanentemente e non potranno essere ripristinati.",
    noAccountExists:
      'Non esiste alcun account per "{emailAddress}".',
  },
  labels: {
    or: "o",
    no: "no",
    yes: "si",
    email: "Email",
    continue: "Continua",
    skip: "Salta",
    save: "Salva",
    password: "Password",
    signInPassword: "Accedi con password",
    signInPasscode: "Accedi con codice di accesso",
    forgotYourPassword: "Password dimenticata?",
    back: "Indietro",
    signInPasskey: "Accedi con passkey",
    registerAuthenticator: "Crea una passkey",
    signIn: "Accedi",
    signUp: "Registrati",
    sendNewPasscode: "Invia nuovo codice",
    passwordRetryAfter: "Riprova tra {passwordRetryAfter}",
    passcodeResendAfter: "Richiedi un unovo codice in {passcodeResendAfter}",
    unverifiedEmail: "non verificato",
    primaryEmail: "principale",
    setAsPrimaryEmail: "Imposta come principale",
    verify: "Verifica",
    delete: "Cancella",
    newEmailAddress: "Nuovo indirizoz email",
    newPassword: "Nuova password",
    rename: "Rinomina",
    newPasskeyName: "Nuovo nome passkey",
    addEmail: "Aggiungi email",
    changePassword: "Cambia password",
    createPasskey: "Crea una passkey",
    webauthnUnsupported: "Le Passkeys non sono supportate dal tuo browser",
    signInWith: "Accedi con {provider}",
    deleteAccount: "Si, cancella questo account.",
  },
  errors: {
    somethingWentWrong:
      "Si è verificato un errore tecnico. Riprova più tardi.",
    requestTimeout: "La richiesta è scaduta.",
    invalidPassword: "Email o password sbagliata.",
    invalidPasscode: "Il codice di accesso inserito non è corretto.",
    passcodeAttemptsReached:
      "Il codice di accesso è stato inserito in modo errato troppe volte. Richiedi un nuovo codice.",
    tooManyRequests:
      "Sono state effettuate troppe richieste. Attenti per ripetere l'operazione richiesta.",
    unauthorized: "La sessione è scaduta. Riprova ad accedere.",
    invalidWebauthnCredential: "La passkey non può più essere utilizzata.",
    passcodeExpired: "Il codice di accesso è scaduto. Richiedi un nuovo codice.",
    userVerification:
      "Verifica utente richiesta. Assicurati che il tuo dispositivo di autenticazione è protetto con un PIN o un codice biometrico.",
    emailAddressAlreadyExistsError: "L'indirizzo email è già stato utilizzato.",
    maxNumOfEmailAddressesReached: "Non è possibile aggiungere ulteriori indirizzi email.",
    thirdPartyAccessDenied:
      "Accesso negato. La richiesta è stata cancellata dall'utente o il provider ha negato l'accesso per altre ragioni.",
    thirdPartyMultipleAccounts:
      "Impossibile identificare l'account. L'indirizzo email è utilizzato in più account.",
    thirdPartyUnverifiedEmail:
      "Verifica email richiesta. Verifica l'indirizzo email utilizzato con il tuo provider.",
  },
};
