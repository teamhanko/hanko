import {
  Fragment,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "preact/compat";
import {
  State,
  HankoError,
  WebauthnSupport,
} from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";
import { useFlowState } from "../hooks/UseFlowState";

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
import Paragraph from "../components/paragraph/Paragraph";

interface Props {
  state: State<"login_init">;
}

type IdentifierTypes = "username" | "email" | "identifier";

const LoginInitPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const {
    init,
    initialComponentName,
    uiState,
    setUIState,
    hidePasskeyButtonOnLogin,
    lastLogin,
  } = useContext(AppContext);

  const [isFlowSwitchLoading, setIsFlowSwitchLoading] =
    useState<boolean>(false);
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
    setIdentifierToUIState(identifier);
    return flowState.actions.continue_with_login_identifier.run({
      [identifierType]: identifier,
    });
  };

  const onRegisterClick = async (event: Event) => {
    event.preventDefault();
    setIsFlowSwitchLoading(true);
    init("registration");
  };

  const onRememberMeChange = async (event: Event) => {
    setRememberMe((prev) => !prev);
    return flowState.actions.remember_me.run({ remember_me: !rememberMe });
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

    const nextState = await flowState.actions.thirdparty_oauth.run({
      provider: name,
      redirect_to: window.location.toString(),
    });

    if (nextState.error) {
      setSelectedThirdPartyProvider(null);
    }

    return nextState;
  };

  const showDivider = useMemo(
    () =>
      !!flowState.actions.webauthn_generate_request_options.enabled ||
      !!flowState.actions.thirdparty_oauth.enabled,
    [flowState.actions],
  );

  const inputs = flowState.actions.continue_with_login_identifier.inputs;

  useEffect(() => {
    const inputs = flowState.actions.continue_with_login_identifier.inputs;
    setIdentifierType(
      inputs?.email ? "email" : inputs?.username ? "username" : "identifier",
    );
  }, [flowState]);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.signIn")}</Headline1>
        <ErrorBox state={flowState} error={thirdPartyError} />
        {inputs ? (
          <Fragment>
            <Form
              flowAction={flowState.actions.continue_with_login_identifier}
              onSubmit={onEmailSubmit}
              maxWidth
            >
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
              <Button>{t("labels.continue")}</Button>
            </Form>
            <Divider hidden={!showDivider}>{t("labels.or")}</Divider>
          </Fragment>
        ) : null}
        {flowState.actions.thirdparty_oauth.enabled
          ? flowState.actions.thirdparty_oauth.inputs.provider.allowed_values?.map(
              (v) => {
                return (
                  <Form
                    key={v.value}
                    flowAction={flowState.actions.thirdparty_oauth}
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
              },
            )
          : null}
        {flowState.actions.webauthn_generate_request_options.enabled &&
        !hidePasskeyButtonOnLogin ? (
          <Form
            flowAction={flowState.actions.webauthn_generate_request_options}
          >
            <Button
              secondary
              title={
                !isWebAuthnSupported ? t("labels.webauthnUnsupported") : null
              }
              disabled={!isWebAuthnSupported}
            >
              {t("labels.signInPasskey")}
            </Button>
          </Form>
        ) : null}
        {flowState.actions.remember_me.enabled && (
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
        <Paragraph center>
          <span>{t("labels.dontHaveAnAccount")}</span>
          <Link
            onClick={onRegisterClick}
            loadingSpinnerPosition={"left"}
            isLoading={isFlowSwitchLoading}
          >
            {t("labels.signUp")}
          </Link>
        </Paragraph>
      </Footer>
    </Fragment>
  );
};

export default LoginInitPage;
