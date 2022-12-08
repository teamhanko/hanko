import { useEffect } from "react";
import { ClientOnly } from "remix-utils";
import { registerHankoAuth } from "~/lib/hanko.client";
import { Hanko } from "@teamhanko/hanko-frontend-sdk";
import styles from "~/styles/todo.css";
import type { LinksFunction } from "@remix-run/node";

export const links: LinksFunction = () => {
  return [{ rel: "stylesheet", href: styles }];
};

export default function Index() {
  const handler = async () => {
    const hanko = new Hanko(window.ENV.HANKO_URL);
    const user = await hanko.user.getCurrent();
    console.log(user);
  };

  useEffect(() => {
    registerHankoAuth({ shadow: true });
    document.addEventListener("hankoAuthSuccess", handler);
    return () => document.removeEventListener("hankoAuthSuccess", handler);
  });

  return (
    <div className="content">
      <ClientOnly fallback={"Loading..."}>
        {() => <hanko-auth lang="en" api={window.ENV.HANKO_URL} />}
      </ClientOnly>
    </div>
  );
}
