import { Fragment } from "preact";
import { useContext, useEffect, useMemo, useState } from "preact/compat";

import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";
import { useFlowState } from "../contexts/FlowState";

import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Button from "../components/form/Button";
import Footer from "../components/wrapper/Footer";
import ErrorBox from "../components/error/ErrorBox";
import Headline1 from "../components/headline/Headline1";
import Link from "../components/link/Link";
import Input from "../components/form/Input";
import { HankoError } from "@teamhanko/hanko-frontend-sdk";
import Divider from "../components/spacer/Divider";

interface Props {
  state: State<"registration_init">;
}

const RegistrationInitPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const {
    init,
    uiState,
    setUIState,
    stateHandler,
    setLoadingAction,
    initialComponentName,
  } = useContext(AppContext);
  const { flowState } = useFlowState(props.state);
  const inputs = flowState.actions.register_login_identifier?.(null).inputs;
  const multipleInputsAvailable = !!(inputs?.email && inputs?.username);
  const [thirdPartyError, setThirdPartyError] = useState<
    HankoError | undefined
  >(undefined);

  const onIdentifierSubmit = async (event: Event) => {
    event.preventDefault();
    setLoadingAction("email-submit");
    const nextState = await flowState.actions
      .register_login_identifier({
        email: uiState.email,
        username: uiState.username,
      })
      .run();
    setLoadingAction(null);
    stateHandler[nextState.name](nextState);
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
    init("login");
  };

  const onThirdpartySubmit = async (event: Event, name: string) => {
    event.preventDefault();
    setLoadingAction("thirdparty-submit");

    const nextState = await flowState.actions
      .thirdparty_oauth({
        provider: name,
        redirect_to: window.location.toString(),
      })
      .run();

    stateHandler[nextState.name](nextState);
  };

  const showDivider = useMemo(
    () => !!flowState.actions.thirdparty_oauth?.(null),
    [flowState.actions],
  );

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
      window.location.pathname + searchParams.toString(),
    );
  }, []);

  return (
    <Fragment>
      <Content>
        <Headline1>{t("headlines.signUp")}</Headline1>
        <ErrorBox state={flowState} error={thirdPartyError} />
        {inputs ? (
          <Fragment>
            <Form onSubmit={onIdentifierSubmit} maxWidth>
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
              <Button uiAction={"email-submit"} autofocus>
                {t("labels.continue")}
              </Button>
            </Form>
            <Divider hidden={!showDivider}>{t("labels.or")}</Divider>
          </Fragment>
        ) : null}
        {flowState.actions.thirdparty_oauth?.(null)
          ? flowState.actions
              .thirdparty_oauth(null)
              .inputs.provider.allowed_values?.map((v) => {
                return (
                  <Form
                    onSubmit={(event) => onThirdpartySubmit(event, v.value)}
                  >
                    <Button
                      uiAction={"thirdparty-submit"}
                      secondary
                      // @ts-ignore
                      icon={v.value}
                    >
                      {t("labels.signInWith", { provider: v.name })}
                    </Button>
                  </Form>
                );
              })
          : null}
      </Content>
      <Footer hidden={initialComponentName !== "auth"}>
        <span hidden />
        <Link
          uiAction={"switch-flow"}
          onClick={onLoginClick}
          loadingSpinnerPosition={"left"}
        >
          {t("labels.alreadyHaveAnAccount")}
        </Link>
      </Footer>
    </Fragment>
  );
};

export default RegistrationInitPage;
