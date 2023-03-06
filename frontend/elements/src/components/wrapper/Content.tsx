import { ComponentChildren } from "preact";

import styles from "./styles.sass";

type Props = {
  children: ComponentChildren;
};

const Content = ({ children }: Props) => {
  return <section className={styles.content}>{children}</section>;
};

export default Content;
