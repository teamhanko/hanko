import { StateUpdater, useContext, useEffect } from "preact/compat";
import { Action } from "@teamhanko/hanko-frontend-sdk";
import { AppContext } from "./AppProvider";

export const useFlowEffects = (
  flowAction: Action<any> | undefined,
  setIsLoading: StateUpdater<boolean>,
) => {
  const { hanko, setUIState } = useContext(AppContext);

  useEffect(
    () =>
      hanko.onFlowBeforeStateChanged((detail) => {
        setUIState((prev) => ({ ...prev, isDisabled: true, error: undefined }));

        if (!flowAction || detail.state.flowName != "login") {
          return;
        }

        const { state } = detail;

        if (state.invokedAction == flowAction.name) {
          setIsLoading(true);
        }
      }),
    [flowAction, hanko, setIsLoading, setUIState],
  );

  useEffect(
    () =>
      hanko.onFlowStateChanged((detail) => {
        if (!flowAction || detail.state.flowName != "login") {
          return;
        }

        setIsLoading(false);
      }),
    [hanko, setIsLoading, flowAction],
  );
};
