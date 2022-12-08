import styles from "~/styles/todo.css";
import type { LinksFunction, LoaderArgs } from "@remix-run/node";
import { json } from "@remix-run/node";
import type { Todos } from "~/lib/todo.server";
import { TodoClient } from "~/lib/todo.server";
import { Form, useLoaderData } from "@remix-run/react";
import { extractHankoCookie, requireHankoId } from "~/lib/auth.server";

export const links: LinksFunction = () => {
  return [{ rel: "stylesheet", href: styles }];
};

export const loader = async ({ request }: LoaderArgs) => {
  await requireHankoId(request);
  const hankoCookie = extractHankoCookie(request);
  const todoClient = new TodoClient(
    process.env.REMIX_APP_TODO_API!,
    hankoCookie
  );
  const todoResponse = await todoClient.listTodos();
  const todos = (await todoResponse.json()) as Todos;
  return json({ todos });
};

export default function Todo() {
  const { todos } = useLoaderData<typeof loader>();

  return (
    <>
      <nav className={"nav"}>
        <Form action="/_api/logout" method="post">
          <button className={"button"}>Logout</button>
        </Form>
      </nav>
      <div className={"content"}>
        <h1 className={"headline"}>Todos</h1>
        {/* <div className={'error'}>{error?.message}</div> */}
        <form
          // onSubmit={addTodo}
          className={"form"}
        >
          <input
            required
            className={"input"}
            type={"text"}
            // value={description}
            // onChange={changeDescription}
          />
          <button type={"submit"} className={"button"}>
            +
          </button>
        </form>
        <div className={"list"}>
          {todos.map((todo, index) => (
            <div className={"item"} key={index}>
              <input
                className={"checkbox"}
                id={todo.todoID}
                type={"checkbox"}
                value={todo.todoID}
                checked={todo.checked}
                // onChange={changeCheckbox}
              />
              <label className={"description"} htmlFor={todo.todoID}>
                {todo.description}
              </label>
              <button
                className={"button"}
                // onClick={() => deleteTodo(todo.todoID!)}
              >
                ×
              </button>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}
