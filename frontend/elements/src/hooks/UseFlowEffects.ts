import { StateUpdater, useContext, useEffect } from "preact/compat";
import { Action } from "@teamhanko/hanko-frontend-sdk";
import { AppContext } from "../contexts/AppProvider";

export const useFlowEffects = (
  flowAction: Action<any> | undefined,
  setIsLoading: StateUpdater<boolean>,
  setIsSuccess: StateUpdater<boolean>,
) => {
  const { hanko, setUIState, isOwnFlow } = useContext(AppContext);

  useEffect(
    () =>
      hanko.onBeforeStateChange(({ state }) => {
        if (!flowAction || !isOwnFlow(state)) {
          return;
        }

        setUIState((prev) => ({ ...prev, isDisabled: true, error: undefined }));
        setIsLoading(state.invokedAction.name == flowAction.name);
      }),
    [flowAction, hanko, isOwnFlow, setIsLoading, setUIState],
  );

  useEffect(
    () =>
      hanko.onAfterStateChange(({ state }) => {
        if (!flowAction || !isOwnFlow(state)) {
          return;
        }
        setIsSuccess(state.previousAction?.name == flowAction.name);
        setIsLoading(false);
      }),
    [hanko, setIsSuccess, setIsLoading, flowAction, isOwnFlow],
  );
};
