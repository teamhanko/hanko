import * as preact from "preact";
import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";

import { HankoError, TechnicalError } from "../../lib/Errors";

import ExclamationMark from "./ExclamationMark";

import styles from "./ErrorMessage.module.css";

type Props = {
  error?: Error;
};

const defaultError = new TechnicalError();

const ErrorMessage = ({ error = defaultError }: Props) => {
  const { t } = useContext(TranslateContext);

  const code = error instanceof HankoError ? error.code : null;

  return (
    <section className={styles.errorMessage} hidden={!error}>
      <span>
        <ExclamationMark />
      </span>
      <span id="errorMessage">
        {code ? t(`errors.${code}`) : error ? error.message : null}
      </span>
    </section>
  );
};

export default ErrorMessage;
