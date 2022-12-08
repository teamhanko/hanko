import { useEffect } from "react";
import { ClientOnly } from "remix-utils";
import { registerHankoAuth } from "~/lib/hanko.client";
import { Hanko } from "@teamhanko/hanko-frontend-sdk";
import styles from "~/styles/todo.css";
import type { ActionArgs, LinksFunction } from "@remix-run/node";
import { redirect } from "@remix-run/node";
import { useFetcher } from "@remix-run/react";

export const links: LinksFunction = () => {
  return [{ rel: "stylesheet", href: styles }];
};

export const action = async ({ request }: ActionArgs) => {
  const formData = await request.formData();
  console.log(Object.fromEntries(formData.entries()));
  return redirect("/todo");
};

export default function Index() {
  const fetcher = useFetcher();

  const handler = async () => {
    const hanko = new Hanko(window.ENV.HANKO_URL);
    const user = await hanko.user.getCurrent();
    const data = { hankoId: user.id, email: user.email };
    fetcher.submit(data, { method: "post" });
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
