import { State, StateName } from "@teamhanko/hanko-frontend-sdk";
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
