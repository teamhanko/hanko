import { h } from "preact";
import styles from "./styles.sass";
import cx from "classnames";

interface Props extends h.JSX.HTMLAttributes<HTMLInputElement> {
  label?: string;
}

const Checkbox = ({ label, ...props }: Props) => {
  return (
    <div className={styles.inputWrapper}>
      <label className={styles.checkboxWrapper}>
        <input
          part={"input checkbox-input"}
          type={"checkbox"}
          aria-label={label}
          className={styles.checkbox}
          {...props}
        />
        <span
          className={cx(styles.label, props.disabled ? styles.disabled : null)}
        >
          {label}
        </span>
      </label>
    </div>
  );
};

export default Checkbox;
