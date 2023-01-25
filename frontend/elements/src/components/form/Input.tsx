import * as preact from "preact";
import { h } from "preact";
import { useEffect, useRef } from "preact/compat";

import styles from "./styles.sass";

interface Props extends h.JSX.HTMLAttributes<HTMLInputElement> {
  label?: string;
}

const Input = ({ label, ...props }: Props) => {
  const ref = useRef(null);

  useEffect(() => {
    const { current: element } = ref;
    if (element && props.autofocus) {
      element.focus();
      element.select();
    }
  }, [props.autofocus]);

  return (
    <div className={styles.inputWrapper}>
      <input
        // @ts-ignore
        part={"input text-input"}
        ref={ref}
        aria-label={props.placeholder}
        className={styles.input}
        {...props}
      />
    </div>
  );
};

export default Input;
