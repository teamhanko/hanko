import * as preact from "preact";
import { useEffect, useMemo, useRef } from "preact/compat";
import cx from "classnames";

import styles from "./Input.module.css";

interface Props {
  name: string;
  index: number;
  focus: boolean;
  digit: string;
  disabled?: boolean;
  onChange: (event: Event) => void;
  onKeyDown: (event: Event) => void;
  onInput: (event: Event) => void;
  onPaste: (event: Event) => void;
  onFocus: (event: Event) => void;
}

const InputPasscodeDigit = ({
  name,
  index,
  focus,
  disabled,
  digit = "",
  onChange,
  onKeyDown,
  onInput,
  onPaste,
  onFocus,
}: Props) => {
  const ref = useRef(null);

  const focusInput = () => {
    const { current: element } = ref;
    if (element) {
      element.focus();
      element.select();
    }
  };

  // Autofocus if it's the first input element
  useEffect(() => {
    if (index === 0) {
      focusInput();
    }
  }, [index, disabled]);

  // Focus the current input element
  useMemo(() => {
    if (focus) {
      focusInput();
    }
  }, [focus]);

  return (
    <input
      name={name + index.toString(10)}
      autoComplete={"off"}
      type={"text"}
      maxLength={1}
      ref={ref}
      disabled={disabled}
      value={digit.charAt(0)}
      onChange={onChange}
      onKeyDown={onKeyDown}
      onInput={onInput}
      onPaste={onPaste}
      onFocus={onFocus}
      required={true}
      className={cx(styles.input, styles.passcodeDigit)}
    />
  );
};

export default InputPasscodeDigit;
