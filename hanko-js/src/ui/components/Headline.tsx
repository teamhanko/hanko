import * as preact from "preact";
import { ComponentChildren } from "preact";

import styles from "./Headline.module.css";

type Props = {
  children: ComponentChildren;
};

const Headline = ({ children }: Props) => {
  return <h1 className={styles.title}>{children}</h1>;
};

export default Headline;
