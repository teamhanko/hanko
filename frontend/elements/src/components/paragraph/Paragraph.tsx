import { ComponentChildren } from "preact";

import styles from "./styles.sass";

type Props = {
  children?: ComponentChildren;
};

const Paragraph = ({ children }: Props) => {
  return (
    <p
      // @ts-ignore
      part={"paragraph"}
      className={styles.paragraph}
    >
      {children || "&nbsp;"}
    </p>
  );
};

export default Paragraph;
