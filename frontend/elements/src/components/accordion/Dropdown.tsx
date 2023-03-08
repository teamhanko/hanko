import { ComponentChildren, Fragment, h } from "preact";
import { StateUpdater } from "preact/compat";

import Accordion from "./Accordion";

interface Props {
  name: string;
  title: string | h.JSX.Element;
  children: ComponentChildren;
  checkedItemIndex?: number;
  setCheckedItemIndex: StateUpdater<number>;
}

const Dropdown = ({
  name,
  title,
  children,
  checkedItemIndex,
  setCheckedItemIndex,
}: Props) => {
  return (
    <Accordion
      dropdown
      name={name}
      columnSelector={() => title}
      contentSelector={() => <Fragment>{children}</Fragment>}
      setCheckedItemIndex={setCheckedItemIndex}
      checkedItemIndex={checkedItemIndex}
      data={[{}]}
    />
  );
};

export default Dropdown;
