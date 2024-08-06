import { State } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/State";
import { StateName } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/state-handling";
import { useEffect, useState } from "preact/compat";

export const useFlowState = <T extends StateName>(
  initialFlowState: State<T>,
) => {
  const [flowState, setFlowState] = useState<State<T>>(initialFlowState);

  useEffect(() => {
    if (initialFlowState) {
      setFlowState(initialFlowState);
    }
  }, [initialFlowState]);

  return { flowState };
};
