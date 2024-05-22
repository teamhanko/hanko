import { ComponentChildren, toChildArray } from "preact";

import styles from "./styles.sass";
import cx from "classnames";

type Props = {
  onSubmit?: (event: Event) => void;
  children: ComponentChildren;
  hidden?: boolean;
  maxWidth?: boolean;
};

const Form = ({ onSubmit, children, hidden, maxWidth }: Props) => {
  return !hidden ? (
    <form onSubmit={onSubmit} className={styles.form}>
      <ul className={styles.ul}>
        {toChildArray(children).map((child, index) => (
          <li
            part={"form-item"}
            className={cx(styles.li, maxWidth ? styles.maxWidth : null)}
            key={index}
          >
            {child}
          </li>
        ))}
      </ul>
    </form>
  ) : null;
};

export default Form;
