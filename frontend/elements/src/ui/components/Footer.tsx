import * as preact from "preact";
import { ComponentChildren } from "preact";

import styles from "./Footer.sass";

interface Props {
  children?: ComponentChildren;
}

const Footer = ({ children }: Props) => {
  return <section className={styles.footer}>{children}</section>;
};

export default Footer;
