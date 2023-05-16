import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { NextPage } from "next";
import { useRouter } from "next/router";
import { TodoClient, Todos } from "../util/TodoClient";
import styles from "../styles/Todo.module.css";
import { Hanko } from "@teamhanko/hanko-elements";
import { SessionExpiredModal } from "../components/SessionExpiredModal";

const hankoAPI = process.env.NEXT_PUBLIC_HANKO_API!;
const todoAPI = process.env.NEXT_PUBLIC_TODO_API!;

const Todo: NextPage = () => {
  const hankoClient = useMemo(() => new Hanko(hankoAPI), []);
  const todoClient = useMemo(() => new TodoClient(todoAPI), []);
  const router = useRouter();

  const [todos, setTodos] = useState<Todos>([]);
  const [description, setDescription] = useState<string>("");
  const [error, setError] = useState<Error | null>(null);
  const modalRef = useRef<HTMLDialogElement>(null);

  const redirectToProfile = () => {
    router.push("/profile").catch(setError)
  }

  const redirectToLogin = useCallback(() => {
    router.push("/").catch(setError)
  }, [router]);

  const addTodo = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const todo = { description, checked: false };

    todoClient
      .addTodo(todo)
      .then((res) => {
        if (res.status === 401) {
          modalRef.current?.showModal();
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
    todoClient
      .listTodos()
      .then((res) => {
        if (res.status === 401) {
          modalRef.current?.showModal();
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
  }, [todoClient]);

  const patchTodo = (id: string, checked: boolean) => {
    todoClient
      .patchTodo(id, checked)
      .then((res) => {
        if (res.status === 401) {
          modalRef.current?.showModal();
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
    todoClient
      .deleteTodo(id)
      .then((res) => {
        if (res.status === 401) {
          modalRef.current?.showModal();
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
    hankoClient.user
      .logout()
      .catch((e) => {
        setError(e);
      });
  };

  const changeDescription = (event: React.ChangeEvent<HTMLInputElement>) => {
    setDescription(event.currentTarget.value);
  };

  const changeCheckbox = (event: React.ChangeEvent<HTMLInputElement>) => {
    const { currentTarget } = event;
    patchTodo(currentTarget.value, currentTarget.checked);
  };

  useEffect(() => listTodos(), [listTodos]);

  useEffect(() => hankoClient.onUserLoggedOut(() => {
    redirectToLogin();
  }), [hankoClient, redirectToLogin]);

  useEffect(() => hankoClient.onSessionNotPresent(() => {
    redirectToLogin();
  }), [hankoClient, redirectToLogin]);

  useEffect(() => hankoClient.onSessionExpired(() => {
    modalRef.current?.showModal();
  }), [hankoClient]);

  return (
    <>
      <SessionExpiredModal ref={modalRef} />
      <nav className={styles.nav}>
        <button onClick={logout} className={styles.button}>
          Logout
        </button>
        <button onClick={redirectToProfile} className={styles.button}>
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
