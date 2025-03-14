import {
  Fragment,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "preact/compat";

import { HankoError, WebauthnSupport } from "@teamhanko/hanko-frontend-sdk";
import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";
import { useFlowState } from "../contexts/FlowState";

import Button from "../components/form/Button";
import Input from "../components/form/Input";
import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Divider from "../components/spacer/Divider";
import ErrorBox from "../components/error/ErrorBox";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";
import Footer from "../components/wrapper/Footer";
import Checkbox from "../components/form/Checkbox";
import Spacer from "../components/spacer/Spacer";

interface Props {
  state: State<"login_init">;
}

type IdentifierTypes = "username" | "email" | "identifier";

const LoginInitPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const {
    init,
    hanko,
    initialComponentName,
    setLoadingAction,
    uiState,
    setUIState,
    stateHandler,
    hidePasskeyButtonOnLogin,
    lastLogin,
  } = useContext(AppContext);

  const [identifierType, setIdentifierType] = useState<IdentifierTypes>(null);
  const [identifier, setIdentifier] = useState<string>(
    uiState.username || uiState.email,
  );
  const { flowState } = useFlowState(props.state);
  const isWebAuthnSupported = WebauthnSupport.supported();
  const [thirdPartyError, setThirdPartyError] = useState<
    HankoError | undefined
  >(undefined);
  const [selectedThirdPartyProvider, setSelectedThirdPartyProvider] = useState<
    string | null
  >(null);
  const [rememberMe, setRememberMe] = useState<boolean>(false);

  const onIdentifierInput = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      const { value } = event.target;
      setIdentifier(value);
      setIdentifierToUIState(value);
    }
  };

  const onEmailSubmit = async (event: Event) => {
    event.preventDefault();

    setLoadingAction("email-submit");

    const nextState = await flowState.actions
      .continue_with_login_identifier({ [identifierType]: identifier })
      .run();

    setIdentifierToUIState(identifier);
    setLoadingAction(null);
    await hanko.flow.run(nextState, stateHandler);
  };

  const onPasskeySubmit = async (event: Event) => {
    event.preventDefault();

    setLoadingAction("passkey-submit");

    const nextState = await flowState.actions
      .webauthn_generate_request_options(null)
      .run();

    await hanko.flow.run(nextState, stateHandler);
  };

  const onRegisterClick = async (event: Event) => {
    event.preventDefault();
    init("registration");
  };

  const onRememberMeChange = async (event: Event) => {
    const nextState = await flowState.actions
      .remember_me({ remember_me: !rememberMe })
      .run();
    setRememberMe((prev) => !prev);
    await hanko.flow.run(nextState, stateHandler);
  };

  const setIdentifierToUIState = (value: string) => {
    const setEmail = () =>
      setUIState((prev) => ({ ...prev, email: value, username: null }));
    const setUsername = () =>
      setUIState((prev) => ({ ...prev, email: null, username: value }));
    switch (identifierType) {
      case "email":
        setEmail();
        break;
      case "username":
        setUsername();
        break;
      case "identifier":
        if (value.match(/^[^@]+@[^@]+\.[^@]+$/)) {
          setEmail();
        } else {
          setUsername();
        }
        break;
    }
  };

  const onThirdpartySubmit = async (event: Event, name: string) => {
    event.preventDefault();
    setSelectedThirdPartyProvider(name);

    const nextState = await flowState.actions
      .thirdparty_oauth({
        provider: name,
        redirect_to: window.location.toString(),
      })
      .run();

    if (nextState.error) {
      setSelectedThirdPartyProvider(null);
    }

    await hanko.flow.run(nextState, stateHandler);
  };

  const showDivider = useMemo(
    () =>
      !!flowState.actions.webauthn_generate_request_options?.(null) ||
      !!flowState.actions.thirdparty_oauth?.(null),
    [flowState.actions],
  );

  const inputs =
    flowState.actions.continue_with_login_identifier?.(null).inputs;

  useEffect(() => {
    const inputs =
      flowState.actions.continue_with_login_identifier?.(null).inputs;
    setIdentifierType(
      inputs?.email ? "email" : inputs?.username ? "username" : "identifier",
    );
  }, [flowState]);

  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);

    if (
      searchParams.get("error") == undefined ||
      searchParams.get("error").length === 0
    ) {
      return;
    }

    let errorCode = "";
    switch (searchParams.get("error")) {
      case "access_denied":
        errorCode = "thirdPartyAccessDenied";
        break;
      default:
        errorCode = "somethingWentWrong";
        break;
    }

    const error: HankoError = {
      name: errorCode,
      code: errorCode,
      message: searchParams.get("error_description"),
    };

    setThirdPartyError(error);

    searchParams.delete("error");
    searchParams.delete("error_description");

    history.replaceState(
      null,
      null,
      window.location.pathname +
        (searchParams.size < 1 ? "" : `?${searchParams.toString()}`),
    );
  }, []);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.signIn")}</Headline1>
        <ErrorBox state={flowState} error={thirdPartyError} />
        {inputs ? (
          <Fragment>
            <Form onSubmit={onEmailSubmit} maxWidth>
              {inputs.email ? (
                <Input
                  type={"email"}
                  autoComplete={"username webauthn"}
                  autoCorrect={"off"}
                  flowInput={inputs.email}
                  onInput={onIdentifierInput}
                  value={identifier}
                  placeholder={t("labels.email")}
                  pattern={"^[^@]+@[^@]+\\.[^@]+$"}
                />
              ) : inputs.username ? (
                <Input
                  type={"text"}
                  autoComplete={"username webauthn"}
                  autoCorrect={"off"}
                  flowInput={inputs.username}
                  onInput={onIdentifierInput}
                  value={identifier}
                  placeholder={t("labels.username")}
                />
              ) : (
                <Input
                  type={"text"}
                  autoComplete={"username webauthn"}
                  autoCorrect={"off"}
                  flowInput={inputs.identifier}
                  onInput={onIdentifierInput}
                  value={identifier}
                  placeholder={t("labels.emailOrUsername")}
                />
              )}
              <Button uiAction={"email-submit"}>{t("labels.continue")}</Button>
            </Form>
            <Divider hidden={!showDivider}>{t("labels.or")}</Divider>
          </Fragment>
        ) : null}
        {flowState.actions.webauthn_generate_request_options?.(null) &&
        !hidePasskeyButtonOnLogin ? (
          <Form onSubmit={(event) => onPasskeySubmit(event)}>
            <Button
              uiAction={"passkey-submit"}
              secondary
              title={
                !isWebAuthnSupported ? t("labels.webauthnUnsupported") : null
              }
              disabled={!isWebAuthnSupported}
              icon={"passkey"}
            >
              {t("labels.signInPasskey")}
            </Button>
          </Form>
        ) : null}
        {flowState.actions.thirdparty_oauth?.(null)
          ? flowState.actions
              .thirdparty_oauth(null)
              .inputs.provider.allowed_values?.map((v) => {
                return (
                  <Form
                    key={v.value}
                    onSubmit={(event) => onThirdpartySubmit(event, v.value)}
                  >
                    <Button
                      isLoading={v.value == selectedThirdPartyProvider}
                      secondary
                      // @ts-ignore
                      icon={
                        v.value.startsWith("custom_")
                          ? "customProvider"
                          : v.value
                      }
                      showLastUsed={
                        lastLogin?.login_method == "third_party" &&
                        lastLogin?.third_party_provider == v.value
                      }
                    >
                      {t("labels.signInWith", { provider: v.name })}
                    </Button>
                  </Form>
                );
              })
          : null}
        {flowState.actions.remember_me?.(null) && (
          <Fragment>
            <Spacer />
            <Checkbox
              required={false}
              type={"checkbox"}
              label={t("labels.staySignedIn")}
              checked={rememberMe}
              onChange={onRememberMeChange}
            />
          </Fragment>
        )}
      </Content>
      <Footer hidden={initialComponentName !== "auth"}>
        <span hidden />
        <Link
          uiAction={"switch-flow"}
          onClick={onRegisterClick}
          loadingSpinnerPosition={"left"}
        >
          {t("labels.dontHaveAnAccount")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default LoginInitPage;
