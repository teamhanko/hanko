import { Translation } from "./translations";

export const fr: Translation = {
  headlines: {
    error: "Une erreur s'est produite",
    loginEmail: "Se connecter ou s'inscrire",
    loginEmailNoSignup: "Se connecter",
    loginFinished: "Connexion réussie",
    loginPasscode: "Entrez le code d'accès",
    loginPassword: "Entrez le mot de passe",
    registerAuthenticator: "Créer une clé d'identification",
    registerConfirm: "Créer un compte ?",
    registerPassword: "Définir un nouveau mot de passe",
    profileEmails: "Adresses e-mail",
    profilePassword: "Mot de passe",
    profilePasskeys: "Clés d'identification",
    isPrimaryEmail: "Adresse e-mail principale",
    setPrimaryEmail: "Définir l'adresse e-mail principale",
    createEmail: "Entrer un nouvel e-mail",
    createUsername: "Entrer un nouveau nom d'utilisateur",
    emailVerified: "Vérifié",
    emailUnverified: "Non vérifié",
    emailDelete: "Supprimer",
    renamePasskey: "Renommer la clé d'identification",
    deletePasskey: "Supprimer la clé d'identification",
    lastUsedAt: "Dernière utilisation le",
    createdAt: "Créé le",
    connectedAccounts: "Comptes connectés",
    deleteAccount: "Supprimer le compte",
    accountNotFound: "Compte non trouvé",
    signIn: "Se connecter",
    signUp: "S'inscrire",
    selectLoginMethod: "Sélectionner la méthode de connexion",
    setupLoginMethod: "Configurer la méthode de connexion",
  },
  texts: {
    enterPasscode:
      'Saisissez le code d\'accès qui a été envoyé à "{emailAddress}".',
    enterPasscodeNoEmail:
      "Entrez le code envoyé à votre adresse e-mail principale.",
    setupPasskey:
      "Connectez-vous à votre compte facilement et en toute sécurité avec une clé d'identification. Remarque : Vos données biométriques sont uniquement stockées sur vos appareils et ne seront jamais partagées avec qui que ce soit.",
    createAccount:
      'Aucun compte n\'existe pour "{emailAddress}". Voulez-vous créer un nouveau compte ?',
    passwordFormatHint:
      "Doit contenir entre {minLength} et {maxLength} caractères.",
    setPrimaryEmail: "Définir cette adresse e-mail comme adresse de contact.",
    isPrimaryEmail:
      "Cette adresse e-mail sera utilisée pour vous contacter si nécessaire.",
    emailVerified: "Cette adresse e-mail a été vérifiée.",
    emailUnverified: "Cette adresse e-mail n'a pas été vérifiée.",
    emailDelete:
      "Si vous supprimez cette adresse e-mail, elle ne pourra plus être utilisée pour vous connecter à votre compte. Les clés d'identification qui ont pu être créées avec cette adresse e-mail resteront intactes.",
    renamePasskey:
      "Définissez un nom pour la clé d'identification qui vous aide à identifier où elle est stockée.",
    deletePasskey:
      "Supprimez cette clé d'identification de votre compte. Notez que la clé d'identification continuera d'exister sur vos appareils et devra également y être supprimée.",
    deleteAccount:
      "Êtes-vous sûr de vouloir supprimer ce compte ? Toutes les données seront supprimées immédiatement et ne pourront pas être récupérées.",
    noAccountExists: 'Aucun compte n\'existe pour "{emailAddress}".',
    selectLoginMethodForFutureLogins:
      "Sélectionnez l'une des méthodes de connexion suivantes à utiliser pour les connexions futures.",
    howDoYouWantToLogin: "Comment souhaitez-vous vous connecter ?",
  },
  labels: {
    or: "ou",
    no: "non",
    yes: "oui",
    email: "E-mail",
    continue: "Continuer",
    skip: "Passer",
    save: "Enregistrer",
    password: "Mot de passe",
    passkey: "Clé d'identification",
    passcode: "Code d'accès",
    signInPassword: "Se connecter avec un mot de passe",
    signInPasscode: "Se connecter avec un code d'accès",
    forgotYourPassword: "Mot de passe oublié ?",
    back: "Retour",
    signInPasskey: "Se connecter avec une clé d'identification",
    registerAuthenticator: "Créer une clé d'identification",
    signIn: "Se connecter",
    signUp: "S'inscrire",
    sendNewPasscode: "Envoyer un nouveau code",
    passwordRetryAfter: "Réessayez dans {passwordRetryAfter}",
    passcodeResendAfter: "Demander un nouveau code dans {passcodeResendAfter}",
    unverifiedEmail: "non vérifiée",
    primaryEmail: "principale",
    setAsPrimaryEmail: "Définir comme principale",
    verify: "Vérifier",
    delete: "Supprimer",
    newEmailAddress: "Nouvelle adresse e-mail",
    newPassword: "Nouveau mot de passe",
    rename: "Renommer",
    newPasskeyName: "Nouveau nom de clé d'identification",
    addEmail: "Ajouter une adresse e-mail",
    createPasskey: "Créer une clé d'identification",
    webauthnUnsupported:
      "Les clés d'identification ne sont pas prises en charge par votre navigateur",
    signInWith: "Se connecter avec {provider}",
    deleteAccount: "Oui, supprimer ce compte.",
    emailOrUsername: "E-mail ou Nom d'utilisateur",
    username: "Nom d'utilisateur",
    optional: "facultatif",
    dontHaveAnAccount: "Vous n'avez pas de compte ?",
    alreadyHaveAnAccount: "Vous avez déjà un compte ?",
    changeUsername: "Changer le nom d'utilisateur",
    setUsername: "Définir le nom d'utilisateur",
    changePassword: "Changer le mot de passe",
    setPassword: "Définir le mot de passe",
  },
  errors: {
    somethingWentWrong:
      "Une erreur technique s'est produite. Veuillez réessayer ultérieurement.",
    requestTimeout: "La demande a expiré.",
    invalidPassword: "Mauvais e-mail ou mot de passe.",
    invalidPasscode: "Le code d'accès fourni n'était pas correct.",
    passcodeAttemptsReached:
      "Le code d'accès a été saisi incorrectement trop de fois. Veuillez demander un nouveau code.",
    tooManyRequests:
      "Trop de demandes ont été effectuées. Veuillez attendre pour répéter l'opération demandée.",
    unauthorized: "Votre session a expiré. Veuillez vous connecter à nouveau.",
    invalidWebauthnCredential:
      "Cette clé d'identification ne peut plus être utilisée.",
    passcodeExpired:
      "Le code d'accès a expiré. Veuillez demander un nouveau code.",
    userVerification:
      "Vérification de l'utilisateur requise. Veuillez vous assurer que votre appareil d'authentification est protégé par un code PIN ou une biométrie.",
    emailAddressAlreadyExistsError: "L'adresse e-mail existe déjà.",
    maxNumOfEmailAddressesReached:
      "Aucune autre adresse e-mail ne peut être ajoutée.",
    thirdPartyAccessDenied:
      "Accès refusé. La demande a été annulée par l'utilisateur ou le fournisseur a refusé l'accès pour d'autres raisons.",
    thirdPartyMultipleAccounts:
      "Impossible d'identifier le compte. L'adresse e-mail est utilisée par plusieurs comptes.",
    thirdPartyUnverifiedEmail:
      "Vérification de l'adresse e-mail requise. Veuillez vérifier l'adresse e-mail utilisée avec votre fournisseur.",
    signupDisabled: "L'enregistrement du compte est désactivé.",
  },
  flowErrors: {
    technical_error:
      "Une erreur technique s'est produite. Veuillez réessayer ultérieurement.",
    flow_expired_error:
      "La session a expiré, veuillez cliquer sur le bouton pour redémarrer.",
    value_invalid_error: "La valeur saisie est invalide.",
    passcode_invalid: "Le code fourni n'était pas correct.",
    passkey_invalid: "Cette clé de passe ne peut plus être utilisée.",
    passcode_max_attempts_reached:
      "Le code a été entré incorrectement trop de fois. Veuillez demander un nouveau code.",
    rate_limit_exceeded:
      "Trop de demandes ont été effectuées. Veuillez patienter pour répéter l'opération demandée.",
    unknown_username_error: "Le nom d'utilisateur est inconnu.",
    username_already_exists: "Le nom d'utilisateur est déjà pris.",
    email_already_exists: "L'adresse e-mail est déjà utilisée.",
    not_found: "La ressource demandée n'a pas été trouvée.",
    operation_not_permitted_error: "L'opération n'est pas autorisée.",
    flow_discontinuity_error:
      "Le processus ne peut pas continuer en raison des paramètres utilisateur ou de la configuration du fournisseur.",
    form_data_invalid_error:
      "Les données du formulaire soumises contiennent des erreurs.",
    unauthorized: "Votre session a expiré. Veuillez vous connecter à nouveau.",
    value_missing_error: "La valeur est manquante.",
    value_too_long_error: "La valeur est trop longue.",
    value_too_short_error: "La valeur est trop courte.",
  },
};
