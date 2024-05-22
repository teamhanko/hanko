import { ComponentChildren, Fragment, h } from "preact";
import { StateUpdater } from "preact/compat";

import Accordion from "./Accordion";

interface Props {
  name: string;
  title: string | h.JSX.Element;
  children: ComponentChildren;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
}

const Dropdown = ({
  name,
  title,
  children,
  checkedItemID,
  setCheckedItemID,
}: Props) => {
  return (
    <Accordion
      dropdown
      name={name}
      columnSelector={() => title}
      contentSelector={() => <Fragment>{children}</Fragment>}
      setCheckedItemID={setCheckedItemID}
      checkedItemID={checkedItemID}
      data={[{}]}
    />
  );
};

export default Dropdown;
