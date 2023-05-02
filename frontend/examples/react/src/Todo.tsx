import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { TodoClient, Todos } from "./TodoClient";
import styles from "./Todo.module.css";
import { createHankoClient } from "@teamhanko/hanko-elements";

const todoAPI = process.env.REACT_APP_TODO_API!;
const hankoAPI = process.env.REACT_APP_HANKO_API!;

function Todo() {
  const navigate = useNavigate();
  const hankoClient = createHankoClient(hankoAPI);
  const [todos, setTodos] = useState<Todos>([]);
  const [description, setDescription] = useState<string>("");
  const [error, setError] = useState<Error | null>(null);
  const todoClient = useMemo(() => new TodoClient(todoAPI), []);

  const redirectToLogin = useCallback(() => {
    navigate("/");
  }, [navigate]);

  const redirectToProfile = () => {
    navigate("/profile");
  }

  const addTodo = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const todo = { description, checked: false };

    todoClient
      .addTodo(todo)
      .then((res) => {
        if (res.status === 401) {
          redirectToLogin();
          return;
        }

        setDescription("");
        listTodos();

        return;
      })
      .catch(setError);
  };

  const listTodos = useCallback(() => {
    todoClient
      .listTodos()
      .then((res) => {
        if (res.status === 401) {
          redirectToLogin();
          return;
        }

        return res.json();
      })
      .then((todo) => {
        if (todo) {
          setTodos(todo);
        }
      })
      .catch(setError);
  }, [todoClient, redirectToLogin]);

  const patchTodo = (id: string, checked: boolean) => {
    todoClient
      .patchTodo(id, checked)
      .then((res) => {
        if (res.status === 401) {
          redirectToLogin();
          return;
        }

        listTodos();

        return;
      })
      .catch(setError);
  };

  const deleteTodo = (id: string) => {
    todoClient
      .deleteTodo(id)
      .then((res) => {
        if (res.status === 401) {
          redirectToLogin();
          return;
        }

        listTodos();

        return;
      })
      .catch(setError);
  };

  const logout = () => {
    hankoClient.user
      .logout()
      .catch(setError);
  };

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

  useEffect(() => hankoClient.onSessionRemoved(() => navigate("/")), [hankoClient, navigate])

  return (
    <>
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
}

export default Todo;
