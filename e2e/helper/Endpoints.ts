const Endpoints = {
  API: {
    ME: "**/me",
    USER: "**/user",
    USERS: "**/users",
    USERS_PARAM: "**/users/*",
    PASSWORD: "**/password",
    PASSWORD_LOGIN: "**/password/login",
    WEBAUTHN_LOGIN_INITIALIZE: "**/webauthn/login/initialize",
    WEBAUTHN_LOGIN_FINALIZE: "**/webauthn/login/finalize",
    WEBAUTHN_REGISTRATION_INITIALIZE: "**/webauthn/registration/initialize",
    WEBAUTHN_REGISTRATION_FINALIZE: "**/webauthn/registration/finalize",
    PASSCODE_LOGIN_INITIALIZE: "**/passcode/login/initialize",
    PASSCODE_LOGIN_FINALIZE: "**/passcode/login/finalize",
    WELL_KNOWN_CONFIG: "**/.well-known/config",
  },
  APP: {
    LOGOUT: "**/logout",
    SECURED_CONTENT: "**/secured",
  },
};

export default Endpoints;
