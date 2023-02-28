import { register } from "@teamhanko/hanko-elements";
import { useCallback, useEffect } from "react";
import { useRouter } from "next/router";

const api = process.env.NEXT_PUBLIC_HANKO_API!;

interface Props {
  setError(error: Error): void;
}

function HankoProfile({ setError }: Props) {
  useEffect(() => {
    register({ shadow: true }).catch(setError);
  }, [setError]);

  return <hanko-profile api={api} />;
}

export default HankoProfile;
