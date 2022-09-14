import * as preact from "preact";
import { useContext, useEffect } from "preact/compat";

import { UnauthorizedError } from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { UserContext } from "../contexts/UserProvider";
import { RenderContext } from "../contexts/PageProvider";

import LoadingIndicator from "../components/LoadingIndicator";

const Initialize = () => {
  const { config, configInitialize } = useContext(AppContext);
  const { userInitialize } = useContext(UserContext);
  const {
    eventuallyRenderEnrollment,
    renderLoginEmail,
    renderLoginFinished,
    renderError,
  } = useContext(RenderContext);

  useEffect(() => {
    configInitialize().catch((e) => renderError(e));
  }, [configInitialize, renderError]);

  useEffect(() => {
    if (config === null) {
      return;
    }

    userInitialize()
      .then((u) => eventuallyRenderEnrollment(u, false))
      .then((rendered) => {
        if (!rendered) {
          renderLoginFinished();
        }

        return;
      })
      .catch((e) => {
        if (e instanceof UnauthorizedError) {
          renderLoginEmail();
        } else {
          renderError(e);
        }
      });
  }, [
    config,
    eventuallyRenderEnrollment,
    renderError,
    renderLoginEmail,
    renderLoginFinished,
    userInitialize,
  ]);

  return <LoadingIndicator isLoading />;
};

export default Initialize;
