import { createHankoClient, register } from "@teamhanko/hanko-elements";
import { useCallback, useEffect, useMemo } from "react";
import { router } from "next/client";

const api = process.env.NEXT_PUBLIC_HANKO_API!;

interface Props {
  setError(error: Error): void;
}

function HankoProfile({ setError }: Props) {
  const hankoClient = useMemo(() => createHankoClient(api), []);

  const redirectToLogin = useCallback(() => {
    router.replace("/").catch(setError);
  }, [setError]);

  useEffect(() => {
    register(api).catch(setError);
  }, [setError]);

  useEffect(() => hankoClient.onSessionRemoved(() => {
    redirectToLogin();
  }))

  return <hanko-profile />;
}

export default HankoProfile;
