import { ComponentChildren } from "preact";

import styles from "./styles.sass";
import cx from "classnames";

type Props = {
  hidden?: boolean;
  children: ComponentChildren;
  center?: boolean;
};

const Paragraph = ({ children, hidden, center }: Props) => {
  return !hidden ? (
    <p
      part={"paragraph"}
      className={cx(
        styles.paragraph,
        center && styles.center,
        center && styles.column,
      )}
    >
      {children}
    </p>
  ) : null;
};

export default Paragraph;
