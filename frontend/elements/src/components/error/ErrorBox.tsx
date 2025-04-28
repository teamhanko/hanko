import { TranslateContext } from "@denysvuika/preact-translate";
import { FlowError, HankoError, State } from "@teamhanko/hanko-frontend-sdk";
import { useContext, useEffect } from "preact/compat";
import { AppContext } from "../../contexts/AppProvider";

import styles from "./styles.sass";
import Icon from "../icons/Icon";

type Props = {
  state?: State<any>;
  flowError?: FlowError;
  error?: HankoError;
};

const ErrorBox = ({ state, error, flowError }: Props) => {
  const { t } = useContext(TranslateContext);
  const { uiState, setUIState } = useContext(AppContext);

  useEffect(() => {
    if (state?.error?.code == "form_data_invalid_error") {
      for (const action of Object.values(state?.actions)) {
        let relatedInputFound = false;
        // @ts-ignore
        for (const input of Object.values(action?.inputs)) {
          // @ts-ignore
          if (input.error?.code) {
            // @ts-ignore
            setUIState({ ...uiState, error: input.error });
            relatedInputFound = true;
            return;
          }
        }

        if (!relatedInputFound) {
          setUIState({ ...uiState, error: state.error });
        }
      }
    } else if (state?.error) {
      setUIState({ ...uiState, error: state?.error });
    }
  }, [state]);

  return (
    <section
      part={"error"}
      className={styles.errorBox}
      hidden={!uiState.error?.code && !flowError?.code && !error}
    >
      <span>
        <Icon name={"exclamation"} size={15} />
      </span>
      <span id="errorMessage" part={"error-text"}>
        {error
          ? t(`errors.${error.code}`)
          : t(`flowErrors.${uiState.error?.code || flowError?.code}`)}
      </span>
    </section>
  );
};

export default ErrorBox;
