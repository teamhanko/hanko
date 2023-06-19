import { register } from "@teamhanko/hanko-elements";
import { useEffect } from "react";

const api = process.env.NEXT_PUBLIC_HANKO_API!;

interface Props {
  setError(error: Error): void;
}

function HankoProfile({ setError }: Props) {
  useEffect(() => {
    register(api).catch(setError);
  }, [setError]);

  return <hanko-profile />;
}

export default HankoProfile;
