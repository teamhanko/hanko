import * as preact from "preact";
import { ComponentChildren, h } from "preact";
import { forwardRef } from "preact/compat";

import styles from "./styles.sass";

interface Props extends h.JSX.HTMLAttributes<HTMLElement> {
  children: ComponentChildren;
}

const Container = forwardRef<HTMLElement>((props: Props, ref) => {
  return (
    <section
      // @ts-ignore
      part={"container"}
      className={styles.container}
      ref={ref}
    >
      {props.children}
    </section>
  );
});

export default Container;
