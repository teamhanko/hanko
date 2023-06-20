import { Translation } from "./translations";

export const fr: Translation = {
  headlines: {
    error: "Une erreur s'est produite",
    loginEmail: "Se connecter ou s'inscrire",
    loginFinished: "Connexion réussie",
    loginPasscode: "Entrez le code d'accès",
    loginPassword: "Entrez le mot de passe",
    registerAuthenticator: "Enregistrer une clé d´identification",
    registerConfirm: "Créer un compte ?",
    registerPassword: "Définir un nouveau mot de passe",
    profileEmails: "Adresses e-mail",
    profilePassword: "Mot de passe",
    profilePasskeys: "Clés d'identification",
    isPrimaryEmail: "Adresse e-mail principale",
    setPrimaryEmail: "Définir l'adresse e-mail principale",
    emailVerified: "Vérifié",
    emailUnverified: "Non vérifié",
    emailDelete: "Supprimer",
    renamePasskey: "Renommer la clé d´identification",
    deletePasskey: "Supprimer la clé d´identification",
    lastUsedAt: "Dernière utilisation le",
    createdAt: "Créé le",
    connectedAccounts: "Comptes connectés",
    deleteAccount: "Supprimer le compte",
  },
  texts: {
    enterPasscode:
      'Saisissez le code d\'accès qui a été envoyé à "{emailAddress}".',
    setupPasskey:
      "Connectez-vous à votre compte facilement et en toute sécurité avec une clé d´identification. Remarque : Vos données biométriques sont uniquement stockées sur vos appareils et ne seront jamais partagées avec qui que ce soit.",
    createAccount:
      'Aucun compte n\'existe pour "{emailAddress}". Voulez-vous créer un nouveau compte ?',
    passwordFormatHint:
      "Doit contenir entre {minLength} et {maxLength} caractères.",
    manageEmails:
      "Vos adresses e-mail sont utilisées pour la communication et l'authentification.",
    changePassword: "Définir un nouveau mot de passe.",
    managePasskeys:
      "Vos clés d'identification vous permettent de vous connecter à ce compte.",
    isPrimaryEmail:
      "Utilisée pour la communication, les codes d'accès et comme nom d'utilisateur pour les clés d'identification. Pour changer l'adresse e-mail principale, ajoutez d'abord une autre adresse e-mail et définissez-la comme principale.",
    setPrimaryEmail:
      "Définissez cette adresse e-mail comme adresse e-mail principale afin qu'elle soit utilisée pour les communications, les codes d'accès et comme nom d'utilisateur pour les clés d'identification.",
    emailVerified: "Cette adresse e-mail a été vérifiée.",
    emailUnverified: "Cette adresse e-mail n'a pas été vérifiée.",
    emailDelete:
      "Si vous supprimez cette adresse e-mail, elle ne pourra plus être utilisée pour vous connecter à votre compte. Les clés d'identification qui ont pu être créées avec cette adresse e-mail resteront intactes.",
    emailDeleteThirdPartyConnection:
      "Si vous supprimez cette adresse e-mail, elle ne pourra plus être utilisée pour se connecter. Vous ne pourrez également plus vous connecter avec ou reconnecter votre compte {provider}. Les clés d'identification qui ont pu être créées avec cette adresse e-mail resteront intactes.",
    emailDeletePrimary:
      "L'adresse e-mail principale ne peut pas être supprimée. Ajoutez d'abord une autre adresse e-mail et définissez-la comme adresse e-mail principale.",
    renamePasskey:
      "Définissez un nom pour la clé d´identification qui vous aide à identifier où elle est stockée.",
    deletePasskey:
      "Supprimez cette clé d´identification de votre compte. Notez que la clé d´identification continuera d'exister sur vos appareils et devra également y être supprimée.",
    deleteAccount:
      "Êtes-vous sûr de vouloir supprimer ce compte ? Toutes les données seront supprimées immédiatement et ne pourront pas être récupérées.",
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
    signInPassword: "Se connecter avec un mot de passe",
    signInPasscode: "Se connecter avec un code d'accès",
    forgotYourPassword: "Mot de passe oublié ?",
    back: "Retour",
    signInPasskey: "Se connecter avec une clé d´identification",
    registerAuthenticator: "Enregistrer une clé d´identification",
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
    newPasskeyName: "Nouveau nom de clé d´identification",
    addEmail: "Ajouter une adresse e-mail",
    changePassword: "Changer le mot de passe",
    addPasskey: "Ajouter une clé d´identification",
    webauthnUnsupported:
      "Les clés d'identification ne sont pas prises en charge par votre navigateur",
    signInWith: "Se connecter avec {provider}",
    deleteAccount: "Oui, supprimer ce compte.",
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
    invalidWebauthnCredential: "Cette clé d´identification ne peut plus être utilisée.",
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
  },
};
