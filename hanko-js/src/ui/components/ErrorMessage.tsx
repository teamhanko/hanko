import * as preact from "preact";
import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import { HankoError, TechnicalError } from "../../lib/Error";

import ExclamationMark from "./ExclamationMark";

import styles from "./ErrorMessage.sass";

type Props = {
  error?: Error;
};

const defaultError = new TechnicalError();

const ErrorMessage = ({ error = defaultError }: Props) => {
  const { t } = useContext(TranslateContext);

  const code = error instanceof HankoError ? error.code : null;

  return (
    <section
      // @ts-ignore
      part={"error"}
      className={styles.errorMessage}
      hidden={!error}
    >
      <span>
        <ExclamationMark />
      </span>
      <span
        id="errorMessage"
        // @ts-ignore
        part={"error-text"}
      >
        {code ? t(`errors.${code}`) : error ? error.message : null}
      </span>
    </section>
  );
};

export default ErrorMessage;
