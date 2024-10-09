import cx from "classnames";
import styles from "./styles.sass";

type Props = {
  value: string;
};

const LastUsed = ({ value, ...props }: Props) => {
  return <span className={cx(styles.lastUsed)}>{value}</span>;
};

export default LastUsed;
