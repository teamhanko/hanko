import * as preact from "preact";
import { ComponentChildren } from "preact";

import styles from "./Headline.sass";

type Props = {
  children: ComponentChildren;
};

const Headline = ({ children }: Props) => {
  return (
    <h1
      // @ts-ignore
      part={"headline"}
      className={styles.title}
    >
      {children}
    </h1>
  );
};

export default Headline;
