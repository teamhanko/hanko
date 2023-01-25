import * as preact from "preact";
import { ComponentChildren } from "preact";

import cx from "classnames";

import styles from "./styles.sass";

type Props = {
  children: ComponentChildren;
};

const Headline2 = ({ children }: Props) => {
  return (
    <h2
      // @ts-ignore
      part={"headline2"}
      className={cx(styles.headline, styles.grade2)}
    >
      {children}
    </h2>
  );
};

export default Headline2;
