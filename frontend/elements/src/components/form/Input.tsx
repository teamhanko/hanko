import { h } from "preact";
import { useContext, useEffect, useMemo, useRef } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { Input as FlowInput } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/input";
import { AppContext } from "../../contexts/AppProvider";
import cx from "classnames";

import styles from "./styles.sass";

interface Props extends h.JSX.HTMLAttributes<HTMLInputElement> {
  label?: string;
  markOptional?: boolean;
  markError?: boolean;
  flowInput?: FlowInput<any>;
}

const Input = ({ label, ...props }: Props) => {
  const ref = useRef(null);
  const { isDisabled } = useContext(AppContext);
  const { t } = useContext(TranslateContext);

  const disabled = useMemo(
    () => isDisabled || props.disabled,
    [props, isDisabled],
  );

  useEffect(() => {
    const { current: element } = ref;
    if (element && props.autofocus) {
      element.focus();
      element.select();
    }
  }, [props.autofocus]);

  const placeholder = useMemo(() => {
    if (props.markOptional && !props.flowInput?.required) {
      return `${props.placeholder} (${t("labels.optional")})`;
    }
    return props.placeholder;
  }, [props.markOptional, props.placeholder, props.flowInput, t]);

  return (
    <div className={styles.inputWrapper}>
      <input
        part={"input text-input"}
        required={props.flowInput?.required}
        maxLength={props.flowInput?.max_length}
        minLength={props.flowInput?.min_length}
        hidden={props.flowInput?.hidden}
        {...props}
        ref={ref}
        aria-label={placeholder}
        placeholder={placeholder}
        className={cx(
          styles.input,
          !!props.flowInput?.error && props.markError && styles.error,
        )}
        disabled={disabled}
      />
    </div>
  );
};

export default Input;
