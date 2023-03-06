import { ComponentChildren } from "preact";

import cx from "classnames";

import styles from "./styles.sass";

type Props = {
  children: ComponentChildren;
};

const Headline1 = ({ children }: Props) => {
  return (
    <h1
      // @ts-ignore
      part={"headline1"}
      className={cx(styles.headline, styles.grade1)}
    >
      {children}
    </h1>
  );
};

export default Headline1;
