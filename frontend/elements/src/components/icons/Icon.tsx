import * as icons from "./icons";

export type IconName = keyof typeof icons;

export type IconProps = {
  secondary?: boolean;
  fadeOut?: boolean;
  disabled?: boolean;
  size?: number;
};

type Props = IconProps & {
  name: IconName;
};

const Icon = ({ name, secondary, size = 18, fadeOut, disabled }: Props) => {
  const Ico = icons[name];

  return (
    <Ico
      size={size}
      secondary={secondary}
      fadeOut={fadeOut}
      disabled={disabled}
    />
  );
};

export default Icon;
