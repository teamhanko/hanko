import { h } from "preact";
import styles from "./styles.sass";
import cx from "classnames";
import { useContext, useMemo } from "preact/compat";
import { AppContext } from "../../contexts/AppProvider";

interface Props extends h.JSX.HTMLAttributes<HTMLInputElement> {
  label?: string;
}

const Checkbox = ({ label, ...props }: Props) => {
  const { uiState } = useContext(AppContext);

  const disabled = useMemo(
    () => uiState.isDisabled || props.disabled,
    [props, uiState],
  );

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
        <span className={cx(styles.label, disabled ? styles.disabled : null)}>
          {label}
        </span>
      </label>
    </div>
  );
};

export default Checkbox;
