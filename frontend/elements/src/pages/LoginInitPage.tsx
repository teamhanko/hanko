import { Fragment, useContext, useEffect, useMemo, useState, } from "preact/compat";

import { WebauthnSupport } from "@teamhanko/hanko-frontend-sdk";
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
import { get as getWebauthnCredential } from "@github/webauthn-json";

interface Props {
  state: State<"login_init">;
}

type IdentifierTypes = "username" | "email" | "identifier";

const LoginInitPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const {
    init,
    initialComponentName,
    setLoadingAction,
    uiState,
    setUIState,
    stateHandler,
    hidePasskeyButtonOnLogin,
  } = useContext(AppContext);

  const [identifierType, setIdentifierType] = useState<IdentifierTypes>(null);
  const { flowState } = useFlowState(props.state);
  const isWebAuthnSupported = WebauthnSupport.supported();
  let controller = new AbortController();

  const onIdentifierInput = (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      const { value } = event.target;

      switch (identifierType) {
        case "email":
          setUIState((prev) => ({ ...prev, email: value, username: null }));
          break;
        case "username":
          setUIState((prev) => ({ ...prev, email: null, username: value }));
          break;
        case "identifier":
          if (value.match(/^[^@]+@[^@]+\.[^@]+$/)) {
            setUIState((prev) => ({ ...prev, email: value, username: null }));
          } else {
            setUIState((prev) => ({ ...prev, email: null, username: value }));
          }
          break;
      }
    }
  };

  const onEmailSubmit = async (event: Event) => {
    event.preventDefault();

    setLoadingAction("email-submit");

    const nextState = await flowState.actions
      .continue_with_login_identifier({
        [identifierType]: uiState.email || uiState.username,
      })
      .run();

    setLoadingAction(null);

    stateHandler[nextState.name](nextState);
  };

  const onPasskeySubmit = async (event: Event) => {
    event.preventDefault();

    // _createAbortSignal()

    setLoadingAction("passkey-submit");
    // await new Promise(r => setTimeout(r, 2000))

    const nextState = await flowState.actions
      .webauthn_generate_request_options(null)
      .run();

    stateHandler[nextState.name](nextState);
  };

  const onRegisterClick = async (event: Event) => {
    event.preventDefault();
    init("registration");
  };

  const showDivider = useMemo(
    () => !!flowState.actions.webauthn_generate_request_options?.(null),
    [flowState.actions],
  );

  const { inputs } = flowState.actions.continue_with_login_identifier(null);

  useEffect(() => {
    const { inputs } = flowState.actions.continue_with_login_identifier(null);
    if (inputs.email) {
      setIdentifierType("email");
    } else if (inputs.username) {
      setIdentifierType("username");
    } else {
      setIdentifierType("identifier");
    }
  }, [flowState]);

  useEffect(() => {
    if (!!flowState.payload.request_options) {
      // conditionalMediationRequest().then(() => {
      // })
    }
    return () => {
      if (controller) {
        console.log("abort webauthn request in cleanup function")
        controller.abort();
      }
    }
  }, []);

  const _createAbortSignal = () => {
    if (controller) {
      console.log("abort webauthn request");
      controller.abort();
      console.log(controller);
    }

    controller = new AbortController();
    console.log(controller);
    return controller.signal;
  };

  const conditionalMediationRequest = async () => {
    const options = { ...flowState.payload.request_options }
    options.signal = _createAbortSignal()
    // `CredentialMediationRequirement` doesn't support "conditional" in the current typescript version.
    options.mediation = "conditional" as CredentialMediationRequirement

    try {
      console.log(options)
      const assertionResponse = await getWebauthnCredential(
        options,
      );

      const nextState = await flowState.actions
        .webauthn_verify_assertion_response({
          assertion_response: assertionResponse,
        })
        .run();

      stateHandler[nextState.name](nextState);
    } catch (error) {
      // We do not need to handle the error, because this is a conditional request, which can fail silently
      console.log(error)
    }
  }

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.signIn")}</Headline1>
        <ErrorBox state={flowState} />
        <Form onSubmit={onEmailSubmit} maxWidth>
          {inputs.email ? (
            <Input
              type={"email"}
              autoComplete={"username webauthn"}
              autoCorrect={"off"}
              flowInput={inputs.email}
              onInput={onIdentifierInput}
              value={uiState.email}
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
              value={uiState.username}
              placeholder={t("labels.username")}
            />
          ) : (
            <Input
              type={"text"}
              autoComplete={"username webauthn"}
              autoCorrect={"off"}
              flowInput={inputs.identifier}
              onInput={onIdentifierInput}
              value={uiState.username || uiState.email}
              placeholder={t("labels.emailOrUsername")}
            />
          )}
          <Button uiAction={"email-submit"}>{t("labels.continue")}</Button>
        </Form>
        <Divider hidden={!showDivider}>{t("labels.or")}</Divider>
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