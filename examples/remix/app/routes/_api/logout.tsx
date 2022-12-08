import { redirect } from "@remix-run/node";
import type { DataFunctionArgs } from "@remix-run/node";
import { serialize } from "cookie";
import { extractHankoCookie } from "~/lib/auth.server";
import { TodoClient } from "~/lib/todo.server";

// This will first log you out of the express backend and then remove the hanko cookie from
// the browser.
export const action = async ({ request }: DataFunctionArgs) => {
  const hankoCookie = extractHankoCookie(request);
  const todoClient = new TodoClient(
    process.env.REMIX_APP_TODO_API!,
    hankoCookie
  );
  await todoClient.logout();
  const loggedOutHankoCookie = serialize("hanko", "", {
    path: "/",
    domain: "localhost",
    maxAge: -1,
    httpOnly: true,
    secure: true,
  });
  return redirect(`/`, { headers: { "Set-Cookie": loggedOutHankoCookie } });
};
