import { ComponentChildren } from "preact";

import styles from "./styles.sass";

type Props = {
  hidden?: boolean;
  children: ComponentChildren;
};

const Paragraph = ({ children, hidden }: Props) => {
  return !hidden ? (
    <p part={"paragraph"} className={styles.paragraph}>
      {children}
    </p>
  ) : null;
};

export default Paragraph;
