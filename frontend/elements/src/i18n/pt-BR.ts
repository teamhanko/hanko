import { Translation } from "./translations";

export const ptBR: Translation = {
  headlines: {
    error: "Ocorreu um erro",
    loginEmail: "Entre ou cadastre-se",
    loginEmailNoSignup: "Entre",
    loginFinished: "Login efetuado com sucesso",
    loginPasscode: "Digite o código de acesso",
    loginPassword: "Digite a senha",
    registerAuthenticator: "Criar uma chave de acesso",
    registerConfirm: "Concluir cadastro?",
    registerPassword: "Redefina sua senha",
    profileEmails: "E-mails",
    profilePassword: "Senha",
    profilePasskeys: "Chave de acesso",
    isPrimaryEmail: "E-mail principal",
    setPrimaryEmail: "Definir o e-mail principal",
    emailVerified: "Verificado",
    emailUnverified: "Não verificado",
    emailDelete: "Apagar",
    renamePasskey: "Renomear a chave de acesso",
    deletePasskey: "Apagar a chave de acesso",
    lastUsedAt: "Usado pela última vez em",
    createdAt: "Criado em",
    connectedAccounts: "Contas conectadas",
    deleteAccount: "Apagar a conta",
    accountNotFound: "Conta não encontrada",
  },
  texts: {
    enterPasscode:
      'Digite o código de acesso que foi enviado para "{emailAddress}".',
    setupPasskey:
      "Entre na sua conta de forma fácil e segura com uma chave de acesso. Nota: Os seus dados biométricos são apenas guardados no seu aparelho e nunca serão compartilhados com ninguém.",
    createAccount:
      'Nenhuma conta encontrada para o e-mail "{emailAddress}". Deseja criar uma nova conta?',
    passwordFormatHint:
      "Deve conter entre {minLength} e {maxLength} caracteres.",
    manageEmails: "Usado para a autenticação com o código de acesso.",
    changePassword: "Alterar senha.",
    managePasskeys:
      "As suas chaves de acesso permitem que você faça login nesta conta.",
    isPrimaryEmail:
      "Este e-mail será usado como seu nome de usuário para as suas chaves de acesso.",
    setPrimaryEmail:
      "Definir este e-mail como nome de usuário para novas chaves de acesso.",
    emailVerified: "Este e-mail foi verificado.",
    emailUnverified: "Este e-mail não foi verificado.",
    emailDelete:
      "Se você apagar esse e-mail, não poderá mais usá-lo para entrar em sua conta.",
    emailDeleteThirdPartyConnection:
      "Se você apagar esse e-mail, não poderá mais usá-lo para entrar em sua conta.",
    emailDeletePrimary: "O seu e-mail principal não pode ser apagado.",
    renamePasskey: "Defina um nome para a chave de acesso.",
    deletePasskey: "Remova essa chave de acesso da sua conta.",
    deleteAccount:
      "Tem certeza que deseja apagar esta conta? Todos os dados serão apagados imediatamente e não poderão ser recuperados.",
    noAccountExists: 'Nenhuma conta encontrada para o e-mail "{emailAddress}".',
  },
  labels: {
    or: "ou",
    no: "não",
    yes: "sim",
    email: "E-mail",
    continue: "Continuar",
    skip: "Pular",
    save: "Salvar",
    password: "Senha",
    signInPassword: "Entre com uma senha",
    signInPasscode: "Entre com um código de acesso",
    forgotYourPassword: "Esqueceu a sua senha?",
    back: "Voltar",
    signInPasskey: "Entre com uma chave de acesso",
    registerAuthenticator: "Criar uma chave de acesso",
    signIn: "Entrar",
    signUp: "Cadastrar-se",
    sendNewPasscode: "Enviar novo código",
    passwordRetryAfter: "Tente novamente em {passwordRetryAfter}",
    passcodeResendAfter: "Peça outro código em {passcodeResendAfter}",
    unverifiedEmail: "Não verificado",
    primaryEmail: "principal",
    setAsPrimaryEmail: "Definir como principal",
    verify: "Verificar",
    delete: "Apagar",
    newEmailAddress: "Novo e-mail",
    newPassword: "Nova senha",
    rename: "Renomear",
    newPasskeyName: "Novo nome para a chave de acesso",
    addEmail: "Adicionar e-mail",
    changePassword: "Mudar a senha",
    createPasskey: "Criar uma chave de acesso",
    webauthnUnsupported:
      "Chaves de acesso não são compatíveis com seu navegador",
    signInWith: "Entre com {provider}",
    deleteAccount: "Sim, apagar esta conta.",
  },
  errors: {
    somethingWentWrong:
      "Ocorreu um erro técnico. Por favor, tente novamente mais tarde.",
    requestTimeout: "A página demorou demais para se conectar.",
    invalidPassword: "E-mail ou senha inválido.",
    invalidPasscode: "O código de acesso inserido não é válido.",
    passcodeAttemptsReached:
      "Um código de acesso inválido foi inserido várias vezes. Por favor, solicite um novo código.",
    tooManyRequests:
      "Muitas tentativas foram feitas. Aguarde alguns minutos antes de tentar novamente.",
    unauthorized: "A sua sessão expirou. Inicie uma nova sessão.",
    invalidWebauthnCredential: "Esta chave de acesso já não pode ser usada.",
    passcodeExpired:
      "O seu código de acesso expirou. Por favor, solicite um novo.",
    userVerification:
      "Verificação de usuário necessária. Por favor, verifique se o seu dispositivo de verificação está protegido com um PIN ou biometria.",
    emailAddressAlreadyExistsError: "Este endereço de e-mail já existe.",
    maxNumOfEmailAddressesReached: "Não é possível adicionar mais e-mails.",
    thirdPartyAccessDenied:
      "Acesso negado. O pedido foi cancelado pelo usuário ou o provedor negou o acesso por outros motivos.",
    thirdPartyMultipleAccounts:
      "Não foi possível identificar a conta. O endereço de e-mail é usado por várias contas.",
    thirdPartyUnverifiedEmail:
      "Verificação de e-mail necessária. Por favor, verifique o e-mail utilizado com o seu provedor.",
  },
};
