import React, { useCallback, useEffect, useMemo, useState } from "react";
import { NextPage } from "next";
import { useRouter } from "next/router";
import { TodoClient, Todos } from "../util/TodoClient";
import styles from "../styles/Todo.module.css";

const api = process.env.NEXT_PUBLIC_TODO_API!;

const Todo: NextPage = () => {
  const client = useMemo(() => new TodoClient(api), []);
  const router = useRouter();

  const [todos, setTodos] = useState<Todos>([]);
  const [description, setDescription] = useState<string>("");
  const [error, setError] = useState<Error | null>(null);

  const addTodo = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const todo = { description, checked: false };

    client
      .addTodo(todo)
      .then((res) => {
        if (res.status === 401) {
          router.replace("/").catch((e) => setError(e));
          return;
        }

        setDescription("");
        listTodos();

        return;
      })
      .catch((e) => {
        setError(e);
      });
  };

  const listTodos = useCallback(() => {
    client
      .listTodos()
      .then((res) => {
        if (res.status === 401) {
          router.push("/").catch((e) => setError(e));
          return;
        }

        return res.json();
      })
      .then((todo) => {
        if (todo) {
          setTodos(todo);
        }
      })
      .catch((e) => {
        setError(e);
      });
  }, [client, router]);

  const patchTodo = (id: string, checked: boolean) => {
    client
      .patchTodo(id, checked)
      .then((res) => {
        if (res.status === 401) {
          router.push("/").catch((e) => setError(e));
          return;
        }

        listTodos();

        return;
      })
      .catch((e) => {
        setError(e);
      });
  };

  const deleteTodo = (id: string) => {
    client
      .deleteTodo(id)
      .then((res) => {
        if (res.status === 401) {
          router.push("/").catch((e) => setError(e));
          return;
        }

        listTodos();

        return;
      })
      .catch((e) => {
        setError(e);
      });
  };

  const logout = () => {
    client
      .logout()
      .then(() => {
        router.push("/").catch((e) => setError(e));
        return;
      })
      .catch((e) => {
        setError(e);
      });
  };

  const profile = () => {
    router.push("/profile").catch(setError)
  }

  const changeDescription = (event: React.ChangeEvent<HTMLInputElement>) => {
    setDescription(event.currentTarget.value);
  };

  const changeCheckbox = (event: React.ChangeEvent<HTMLInputElement>) => {
    const { currentTarget } = event;
    patchTodo(currentTarget.value, currentTarget.checked);
  };

  useEffect(() => {
    listTodos();
  }, [listTodos]);

  return (
    <>
      <nav className={styles.nav}>
        <button onClick={logout} className={styles.button}>
          Logout
        </button>
        <button onClick={profile} className={styles.button}>
          Profile
        </button>
        <button disabled className={styles.button}>
          Todos
        </button>
      </nav>
      <div className={styles.content}>
        <h1 className={styles.headline}>Todos</h1>
        <div className={styles.error}>{error?.message}</div>
        <form onSubmit={addTodo} className={styles.form}>
          <input
            required
            className={styles.input}
            type={"text"}
            value={description}
            onChange={changeDescription}
          />
          <button type={"submit"} className={styles.button}>
            +
          </button>
        </form>
        <div className={styles.list}>
          {todos.map((todo, index) => (
            <div className={styles.item} key={index}>
              <input
                className={styles.checkbox}
                id={todo.todoID}
                type={"checkbox"}
                value={todo.todoID}
                checked={todo.checked}
                onChange={changeCheckbox}
              />
              <label className={styles.description} htmlFor={todo.todoID}>
                {todo.description}
              </label>
              <button className={styles.button} onClick={() => deleteTodo(todo.todoID!)}>
                ×
              </button>
            </div>
          ))}
        </div>
      </div>
    </>
  );
};

export default Todo;
