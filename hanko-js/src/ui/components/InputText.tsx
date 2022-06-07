import * as preact from "preact";
import { useEffect, useRef } from "preact/compat";

import styles from "./Input.module.css";

type Props = {
  name: string;
  type: string;
  value?: string;
  required?: boolean;
  disabled?: boolean;
  label: string;
  onInput: (event: Event) => void;
  autofocus?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
};

const InputText = ({
  name,
  type,
  value,
  required,
  disabled,
  label,
  onInput,
  autofocus,
  minLength,
  maxLength,
  pattern,
}: Props) => {
  const ref = useRef(null);

  useEffect(() => {
    const { current: element } = ref;
    if (element && autofocus) {
      element.focus();
      element.select();
    }
  }, [autofocus, disabled]);

  return (
    <div className={styles.inputWrapper}>
      <input
        ref={ref}
        name={name}
        type={type}
        required={required}
        disabled={disabled}
        onInput={onInput}
        value={value}
        className={styles.input}
        placeholder={" "}
        autofocus={autofocus}
        minLength={minLength}
        maxLength={maxLength}
        pattern={pattern}
      />
      <label className={styles.label}>{label}</label>
    </div>
  );
};

export default InputText;
