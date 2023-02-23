import { register } from "@teamhanko/hanko-elements";
import { useCallback, useEffect } from "react";
import { useRouter } from "next/router";

const api = process.env.NEXT_PUBLIC_HANKO_API!;

function HankoAuth() {
  const router = useRouter();

  const redirectToTodos = useCallback(() => {
    router.replace("/todo");
  }, [router]);

  useEffect(() => {
    register({ shadow: true }).catch((e) => console.error(e));
  }, []);

  useEffect(() => {
    document.addEventListener("hankoAuthSuccess", redirectToTodos);
    return () =>
      document.removeEventListener("hankoAuthSuccess", redirectToTodos);
  }, [redirectToTodos]);

  return <hanko-auth api={api} />;
}

export default HankoAuth;
