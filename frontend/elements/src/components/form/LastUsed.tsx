import { ComponentChildren } from "preact";
import styles from "./styles.sass";

interface Props {
  children: ComponentChildren;
}

const LastUsed = (props: Props) => {
  return <div className={styles.lastUsed}>{props.children}</div>;
};

export default LastUsed;
