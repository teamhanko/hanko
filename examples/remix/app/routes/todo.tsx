import styles from "~/styles/todo.css";
import { redirect } from "@remix-run/node";
import { json } from "@remix-run/node";
import { TodoClient } from "~/lib/todo.server";
import { Form, useLoaderData } from "@remix-run/react";
import { extractHankoCookie, requireValidJwt } from "~/lib/auth.server";
import { badRequest } from "remix-utils";

import type { Todos } from "~/lib/todo.server";
import type { ActionArgs, LinksFunction, LoaderArgs } from "@remix-run/node";

export const links: LinksFunction = () => {
  return [{ rel: "stylesheet", href: styles }];
};

export const loader = async ({ request }: LoaderArgs) => {
  await requireValidJwt(request);
  const hankoCookie = extractHankoCookie(request);
  const todoClient = new TodoClient(
    process.env.REMIX_APP_TODO_API!,
    hankoCookie
  );
  const todoResponse = await todoClient.listTodos();
  const todos = (await todoResponse.json()) as Todos;
  return json({ todos });
};

export const action = async ({ request }: ActionArgs) => {
  const hankoCookie = extractHankoCookie(request);
  const todoClient = new TodoClient(
    process.env.REMIX_APP_TODO_API!,
    hankoCookie
  );
  const formData = await request.formData();
  if (request.method === "POST") {
    const todo = {
      checked: false,
      description: formData.get("description") as string,
    };
    try {
      await todoClient.addTodo(todo);
    } catch (e) {
      return badRequest({ error: e });
    }
  } else {
    const action = formData.get("action") as string;
    const todoID = formData.get("todoID") as string;
    if (action === "delete") {
      try {
        await todoClient.deleteTodo(todoID);
      } catch (e) {
        return badRequest({ error: e });
      }
    } else if (action === "update") {
      const checked = formData.get("checked") === "on";
      try {
        await todoClient.patchTodo(todoID, checked);
      } catch (e) {
        return badRequest({ error: e });
      }
    }
  }
  return redirect(`/todo`);
};

export default function Todo() {
  const { todos } = useLoaderData<typeof loader>();

  return (
    <>
      <nav className="nav">
        <Form action="/_api/logout" method="post">
          <button className="button">Logout</button>
        </Form>
      </nav>
      <div className="content">
        <h1 className="headline">Todos</h1>
        <Form method="post" className="form">
          <input required className="input" type="text" name="description" />
          <button type="submit" className="button">
            +
          </button>
        </Form>
        <div className="list">
          {todos.map((todo, index) => (
            <Form className="item" key={index} method="put">
              <button
                type="submit"
                name="action"
                value="update"
                style={{ all: "unset", cursor: "pointer" }}
              >
                <input
                  className="checkbox"
                  id={todo.todoID}
                  type="checkbox"
                  name="checked"
                  defaultChecked={todo.checked}
                />
              </button>
              <input type="hidden" name="todoID" value={todo.todoID} />
              <label className="description">{todo.description}</label>
              <button
                type="submit"
                name="action"
                className="button"
                value="delete"
              >
                ×
              </button>
            </Form>
          ))}
        </div>
      </div>
    </>
  );
}
