import * as preact from "preact";
import { useEffect, useRef } from "preact/compat";
import { ComponentChildren } from "preact";

import styles from "./Container.sass";

type Props = {
  emitSuccessEvent?: boolean;
  children: ComponentChildren;
};

const Container = ({ children, emitSuccessEvent }: Props) => {
  const ref = useRef(null);

  useEffect(() => {
    if (!emitSuccessEvent) {
      return;
    }

    const event = new Event("success", {
      bubbles: false,
      composed: true,
    });

    const fn = setTimeout(() => {
      ref.current.dispatchEvent(event);
    }, 500);

    return () => clearTimeout(fn);
  }, [emitSuccessEvent]);

  return (
    <section
      // @ts-ignore
      part={"container"}
      ref={ref}
      className={styles.container}
    >
      {children}
    </section>
  );
};

export default Container;
