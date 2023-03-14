import { ComponentChildren, toChildArray } from "preact";

import styles from "./styles.sass";

type Props = {
  onSubmit?: (event: Event) => void;
  children: ComponentChildren;
};

const Form = ({ onSubmit, children }: Props) => {
  return (
    <form onSubmit={onSubmit} className={styles.form}>
      <ul className={styles.ul}>
        {toChildArray(children).map((child, index) => (
          <li part={"form-item"} className={styles.li} key={index}>
            {child}
          </li>
        ))}
      </ul>
    </form>
  );
};

export default Form;
