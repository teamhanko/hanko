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
    otpSetUp: "Configurar o aplicativo de autenticação",
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
    mfaSetUp: "Configurar MFA",
    securityKeySetUp: "Adicionar uma chave de segurança",
    securityKeyLogin: "Chave de segurança",
    otpLogin: "Código de autenticação",
    renameSecurityKey: "Renomear chave de segurança",
    deleteSecurityKey: "Excluir chave de segurança",
    securityKeys: "Chaves de segurança",
    authenticatorApp: "Aplicativo de autenticação",
    authenticatorAppAlreadySetUp:
      "O aplicativo de autenticação já está configurado",
    authenticatorAppNotSetUp: "Configurar o aplicativo de autenticação",
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
    otpEnterVerificationCode:
      "Insira o código de verificação gerado pelo seu aplicativo de autenticação:",
    otpScanQRCode:
      "Escaneie o código QR com seu aplicativo de autenticação (como Google Authenticator ou outro aplicativo TOTP). Alternativamente, você pode inserir manualmente a chave secreta OTP no aplicativo.",
    otpSecretKey: "Chave secreta OTP",
    passwordFormatHint:
      "Deve conter entre {minLength} e {maxLength} caracteres.",
    setPrimaryEmail:
      "Defina este endereço de e-mail para ser usado para entrar em contato com você.",
    isPrimaryEmail:
      "Este endereço de e-mail será usado para entrar em contato com você, se necessário.",
    emailVerified: "Este e-mail foi verificado.",
    emailUnverified: "Este e-mail não foi verificado.",
    emailDelete:
      "Se você apagar esse e-mail, não poderá mais usá-lo para entrar em sua conta.",
    renamePasskey: "Defina um nome para a chave de acesso.",
    deletePasskey: "Remova essa chave de acesso da sua conta.",
    deleteAccount:
      "Tem certeza que deseja apagar esta conta? Todos os dados serão apagados imediatamente e não poderão ser recuperados.",
    noAccountExists: 'Nenhuma conta encontrada para o e-mail "{emailAddress}".',
    selectLoginMethodForFutureLogins:
      "Selecione um dos métodos de login a seguir para usar em logins futuros.",
    howDoYouWantToLogin: "Como você deseja fazer login?",
    mfaSetUp:
      "Proteja sua conta com autenticação de múltiplos fatores (MFA). A MFA adiciona uma camada extra de segurança ao seu processo de login e garante que sua conta permaneça protegida mesmo que sua senha ou endereço de e-mail sejam comprometidos.",
    securityKeyLogin:
      "Conecte sua chave de segurança ou ative-a, em seguida, clique no botão abaixo. Quando estiver pronto, use-a via USB, NFC ou seu telefone. Siga as instruções para concluir o processo de login.",
    otpLogin:
      "Abra seu aplicativo de autenticação para obter o código OTP. Insira o código no campo abaixo para concluir seu login.",
    renameSecurityKey: "Defina um nome para a chave de segurança.",
    deleteSecurityKey: "Exclua esta chave de segurança da sua conta.",
    authenticatorAppAlreadySetUp:
      "Sua conta está protegida por um aplicativo de autenticação que gera códigos únicos (TOTP) para autenticação de múltiplos fatores.",
    authenticatorAppNotSetUp:
      "Proteja sua conta com um aplicativo de autenticação que gera códigos únicos (TOTP) para autenticação de múltiplos fatores.",
    securityKeySetUp:
      "Use uma chave de segurança dedicada via USB, Bluetooth ou NFC ou seu telefone. Conecte sua chave de segurança ou ative-a, em seguida, clique no botão abaixo e siga as instruções para concluir o registro.",
  },
  labels: {
    or: "ou",
    no: "não",
    yes: "sim",
    email: "E-mail",
    continue: "Continuar",
    copied: "copiado",
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
    authenticatorApp: "Aplicativo de autenticação",
    securityKey: "Chave de segurança",
    securityKeyUse: "Usar chave de segurança",
    newSecurityKeyName: "Novo nome da chave de segurança",
    createSecurityKey: "Criar chave de segurança",
    authenticatorAppManage: "Gerenciar aplicativo de autenticação",
    authenticatorAppAdd: "Configurar",
    configured: "configurado",
    useAnotherMethod: "Usar outro método",
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
    handlerNotFoundError:
      "O passo atual não é suportado nesta versão do aplicativo. Por favor, tente novamente mais tarde ou entre em contato com a equipe de suporte para obter ajuda.",
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
    invalid_username_error:
      "O nome de usuário deve conter apenas letras, números e sublinhados.",
    email_already_exists: "O e-mail já está em uso.",
    not_found: "O recurso solicitado não foi encontrado.",
    operation_not_permitted_error: "A operação não é permitida.",
    flow_discontinuity_error:
      "O processo não pode ser continuado devido às configurações do usuário ou do provedor.",
    form_data_invalid_error: "Os dados do formulário submetido contêm erros.",
    unauthorized: "A sua sessão expirou. Inicie uma nova sessão.",
    value_missing_error: "O valor está ausente.",
    value_too_long_error: "O valor é muito longo.",
    value_too_short_error: "O valor é muito curto.",
    webauthn_credential_invalid_mfa_only:
      "Esta identidade pode ser usada apenas como segundo fator de autenticação.",
    webauthn_credential_already_exists:
      "A solicitação expirou, foi cancelada ou o dispositivo já está registrado. Tente novamente ou use outro dispositivo.",
    platform_authenticator_required:
      "Sua conta está configurada para usar autentificadores de plataforma. No entanto, seu dispositivo ou navegador atual não suporta esse recurso. Tente novamente com um dispositivo ou navegador compatível.",
  },
};
