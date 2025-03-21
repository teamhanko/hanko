import { StateUpdater, useContext, useEffect } from "preact/compat";
import { Action } from "@teamhanko/hanko-frontend-sdk";
import { AppContext } from "./AppProvider";

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
        setIsLoading(state.invokedActionID == flowAction.id);
      }),
    [flowAction, hanko, isOwnFlow, setIsLoading, setUIState],
  );

  useEffect(
    () =>
      hanko.onAfterStateChange(({ state }) => {
        if (!flowAction || !isOwnFlow(state)) {
          return;
        }

        setIsSuccess(state.previousActionID == flowAction.id);
        setIsLoading(false);
      }),
    [hanko, setIsSuccess, setIsLoading, flowAction, isOwnFlow],
  );
};
