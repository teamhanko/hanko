import { Hanko, register } from "@teamhanko/hanko-elements";
import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/router";

const api = process.env.NEXT_PUBLIC_HANKO_API!;

interface Props {
  setError(error: Error): void;
}

function HankoAuth({ setError }: Props) {
  const router = useRouter();
  const [hankoClient, setHankoClient] = useState<Hanko>();

  const redirectToTodos = useCallback(() => {
    router.replace("/todo").catch(setError);
  }, [router, setError]);

  useEffect(() => setHankoClient(new Hanko(api)), []);

  useEffect(() => {
    register(api).catch(setError);
  }, [setError]);

  useEffect(() => hankoClient?.onSessionCreated(() => {
    redirectToTodos()
  }), [hankoClient, redirectToTodos]);

  return <hanko-auth />;
}

export default HankoAuth;
