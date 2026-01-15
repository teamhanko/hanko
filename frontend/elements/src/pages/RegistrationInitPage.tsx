import { useContext, useMemo, useState } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import {
  HankoError,
  State,
  generateCodeVerifier,
  setStoredCodeVerifier, clearStoredCodeVerifier
} from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { useFlowState } from "../hooks/UseFlowState";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import Footer from "../components/wrapper/Footer";
import ErrorBox from "../components/error/ErrorBox";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";
import Input from "../components/form/Input";
import Divider from "../components/spacer/Divider";
import Checkbox from "../components/form/Checkbox";
import Spacer from "../components/spacer/Spacer";
import Paragraph from "../components/paragraph/Paragraph";

interface Props {
  state: State<"registration_init">;
}

const RegistrationInitPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const { init, uiState, setUIState, initialComponentName } =
    useContext(AppContext);
  const { flowState } = useFlowState(props.state);
  const inputs = flowState.actions.register_login_identifier.inputs;
  const multipleInputsAvailable = !!(inputs?.email && inputs?.username);
  const [thirdPartyError, setThirdPartyError] = useState<
    HankoError | undefined
  >(undefined);
  const [selectedThirdPartyProvider, setSelectedThirdPartyProvider] = useState<
    string | null
  >(null);
  const [rememberMe, setRememberMe] = useState<boolean>(false);
  const [isFlowSwitchLoading, setIsFlowSwitchLoading] =
    useState<boolean>(false);

  const onIdentifierSubmit = async (event: Event) => {
    event.preventDefault();
    return await flowState.actions.register_login_identifier.run({
      email: uiState.email,
      username: uiState.username,
    });
  };

  const onUsernameInput = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      const { value } = event.target;
      setUIState((prev) => ({ ...prev, username: value }));
    }
  };

  const onEmailInput = (event: Event) => {
    event.preventDefault();
    if (event.target instanceof HTMLInputElement) {
      const { value } = event.target;
      setUIState((prev) => ({ ...prev, email: value }));
    }
  };

  const onLoginClick = async (event: Event) => {
    event.preventDefault();
    setIsFlowSwitchLoading(true);
    init("login");
  };

  const onThirdpartySubmit = async (event: Event, name: string) => {
    event.preventDefault();
    setSelectedThirdPartyProvider(name);

    const codeVerifier = generateCodeVerifier();
    setStoredCodeVerifier(codeVerifier);

    try {
      const nextState = await flowState.actions.thirdparty_oauth.run(
        {
          provider: name,
          redirect_to: window.location.toString(),
          code_verifier: codeVerifier,
        },
        { dispatchAfterStateChangeEvent: false },
      );

      if (nextState.error) {
        clearStoredCodeVerifier();
        setSelectedThirdPartyProvider(null);
      }

      nextState.dispatchAfterStateChangeEvent();
    } catch (e) {
      clearStoredCodeVerifier();
      setSelectedThirdPartyProvider(null);
      throw e;
    }
  };

  const onRememberMeChange = async (event: Event) => {
    event.preventDefault();
    const nextState = await flowState.actions.remember_me.run(
      { remember_me: !rememberMe },
      { dispatchAfterStateChangeEvent: false },
    );
    setRememberMe((prev) => !prev);
    nextState.dispatchAfterStateChangeEvent();
  };

  const showDivider = useMemo(
    () =>
      !!flowState.actions.thirdparty_oauth.enabled &&
      flowState.actions.register_login_identifier.enabled,
    [flowState.actions],
  );

  return (
    <>
      <Content>
        <Headline1>{t("headlines.signUp")}</Headline1>
        <ErrorBox state={flowState} error={thirdPartyError} />
        {inputs ? (
          <>
            <Form
              flowAction={flowState.actions.register_login_identifier}
              onSubmit={onIdentifierSubmit}
              maxWidth
            >
              {inputs.username ? (
                <Input
                  markOptional={multipleInputsAvailable}
                  markError={multipleInputsAvailable}
                  type={"text"}
                  autoComplete={"username"}
                  autoCorrect={"off"}
                  flowInput={inputs.username}
                  onInput={onUsernameInput}
                  value={uiState.username}
                  placeholder={t("labels.username")}
                />
              ) : null}
              {inputs.email ? (
                <Input
                  markOptional={multipleInputsAvailable}
                  markError={multipleInputsAvailable}
                  type={"email"}
                  autoComplete={"email"}
                  autoCorrect={"off"}
                  flowInput={inputs.email}
                  onInput={onEmailInput}
                  value={uiState.email}
                  placeholder={t("labels.email")}
                  pattern={"^.*[^0-9]+$"}
                />
              ) : null}
              <Button autofocus>{t("labels.continue")}</Button>
            </Form>
            <Divider hidden={!showDivider}>{t("labels.or")}</Divider>
          </>
        ) : null}
        {flowState.actions.thirdparty_oauth.enabled
          ? flowState.actions.thirdparty_oauth.inputs.provider.allowed_values?.map(
              (v) => {
                return (
                  <Form
                    flowAction={flowState.actions.thirdparty_oauth}
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
                    >
                      {t("labels.signInWith", { provider: v.name })}
                    </Button>
                  </Form>
                );
              },
            )
          : null}
        {flowState.actions.remember_me.enabled && (
          <>
            <Spacer />
            <Checkbox
              required={false}
              type={"checkbox"}
              label={t("labels.staySignedIn")}
              checked={rememberMe}
              onChange={onRememberMeChange}
            />
          </>
        )}
      </Content>
      <Footer hidden={initialComponentName !== "auth"}>
        <Paragraph center>
          <span>{t("labels.alreadyHaveAnAccount")}</span>
          <Link
            onClick={onLoginClick}
            loadingSpinnerPosition={"left"}
            isLoading={isFlowSwitchLoading}
          >
            {t("labels.signIn")}
          </Link>
        </Paragraph>
      </Footer>
    </>
  );
};

export default RegistrationInitPage;
