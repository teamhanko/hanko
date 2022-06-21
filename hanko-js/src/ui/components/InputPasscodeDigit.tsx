import * as preact from "preact";
import { h } from "preact";
import { useEffect, useMemo, useRef } from "preact/compat";

import styles from "./Input.sass";

interface Props extends h.JSX.HTMLAttributes<HTMLInputElement> {
  index: number;
  focus: boolean;
  digit: string;
}

const InputPasscodeDigit = ({ index, focus, digit = "", ...props }: Props) => {
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
  }, [index, props.disabled]);

  // Focus the current input element
  useMemo(() => {
    if (focus) {
      focusInput();
    }
  }, [focus]);

  return (
    <div className={styles.passcodeDigitWrapper}>
      <input
        {...props}
        name={props.name + index.toString(10)}
        autoComplete={"off"}
        type={"text"}
        maxLength={1}
        ref={ref}
        value={digit.charAt(0)}
        required={true}
        className={styles.input}
      />
    </div>
  );
};

export default InputPasscodeDigit;
