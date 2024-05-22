import { ComponentChildren } from "preact";

import styles from "./styles.sass";

interface Props {
  hidden?: boolean;
  children?: ComponentChildren;
}

const Footer = ({ children, hidden = false }: Props) => {
  return !hidden ? (
    <section className={styles.footer}>{children}</section>
  ) : null;
};

export default Footer;
