import styles from "./styles.sass";
import { Fragment, useContext } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { Error as FlowError } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/error";

interface Props {
  flowError?: FlowError;
}

const ErrorMessage = ({ flowError }: Props) => {
  const { t } = useContext(TranslateContext);
  return (
    <Fragment>
      {flowError ? (
        <div className={styles.errorMessage}>
          {t(`flowErrors.${flowError?.code}`)}
        </div>
      ) : null}
    </Fragment>
  );
};

export default ErrorMessage;
