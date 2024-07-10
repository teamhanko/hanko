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
    createEmail: "Digite um novo e-mail",
    createUsername: "Digite um novo nome de usuário",
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
    signIn: "Entrar",
    signUp: "Registrar",
    selectLoginMethod: "Selecionar método de login",
    setupLoginMethod: "Configurar método de login",
  },
  texts: {
    enterPasscode:
      'Digite o código de acesso que foi enviado para "{emailAddress}".',
    enterPasscodeNoEmail:
      "Digite o código enviado para o seu endereço de e-mail principal.",
    setupPasskey:
      "Entre na sua conta de forma fácil e segura com uma chave de acesso. Nota: Os seus dados biométricos são apenas guardados no seu aparelho e nunca serão compartilhados com ninguém.",
    createAccount:
      'Nenhuma conta encontrada para o e-mail "{emailAddress}". Deseja criar uma nova conta?',
    passwordFormatHint:
      "Deve conter entre {minLength} e {maxLength} caracteres.",
    isPrimaryEmail:
      "Este e-mail será usado como seu nome de usuário para as suas chaves de acesso.",
    setPrimaryEmail:
      "Definir este e-mail como nome de usuário para novas chaves de acesso.",
    emailVerified: "Este e-mail foi verificado.",
    emailUnverified: "Este e-mail não foi verificado.",
    emailDelete:
      "Se você apagar esse e-mail, não poderá mais usá-lo para entrar em sua conta.",
    emailDeletePrimary: "O seu e-mail principal não pode ser apagado.",
    renamePasskey: "Defina um nome para a chave de acesso.",
    deletePasskey: "Remova essa chave de acesso da sua conta.",
    deleteAccount:
      "Tem certeza que deseja apagar esta conta? Todos os dados serão apagados imediatamente e não poderão ser recuperados.",
    noAccountExists: 'Nenhuma conta encontrada para o e-mail "{emailAddress}".',
    selectLoginMethodForFutureLogins:
      "Selecione um dos métodos de login a seguir para usar em logins futuros.",
    howDoYouWantToLogin: "Como você deseja fazer login?",
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
    passkey: "Chave de acesso",
    passcode: "Código de acesso",
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
    createPasskey: "Criar uma chave de acesso",
    webauthnUnsupported:
      "Chaves de acesso não são compatíveis com seu navegador",
    signInWith: "Entre com {provider}",
    deleteAccount: "Sim, apagar esta conta.",
    emailOrUsername: "E-mail ou Nome de usuário",
    username: "Nome de usuário",
    optional: "opcional",
    dontHaveAnAccount: "Não tem uma conta?",
    alreadyHaveAnAccount: "Já tem uma conta?",
    changeUsername: "Alterar nome de usuário",
    setUsername: "Definir nome de usuário",
    changePassword: "Alterar senha",
    setPassword: "Definir senha",
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
    signupDisabled: "O registro da conta está desativado.",
  },
  flowErrors: {
    technical_error:
      "Ocorreu um erro técnico. Por favor, tente novamente mais tarde.",
    flow_expired_error:
      "A sessão expirou, por favor, clique no botão para reiniciar.",
    value_invalid_error: "O valor inserido é inválido.",
    passcode_invalid: "O código fornecido não estava correto.",
    passkey_invalid: "Esta chave de acesso não pode mais ser utilizada.",
    passcode_max_attempts_reached:
      "O código foi inserido incorretamente várias vezes. Por favor, solicite um novo código.",
    rate_limit_exceeded:
      "Foram feitas muitas solicitações. Por favor, aguarde para repetir a operação solicitada.",
    unknown_username_error: "O nome de usuário é desconhecido.",
    username_already_exists: "O nome de usuário já está em uso.",
    email_already_exists: "O e-mail já está em uso.",
    not_found: "O recurso solicitado não foi encontrado.",
    operation_not_permitted_error: "A operação não é permitida.",
    flow_discontinuity_error:
      "O processo não pode ser continuado devido às configurações do usuário ou do provedor.",
    form_data_invalid_error: "Os dados do formulário submetido contêm erros.",
    value_missing_error: "O valor está ausente.",
    value_too_long_error: "O valor é muito longo.",
    value_too_short_error: "O valor é muito curto.",
  },
};
