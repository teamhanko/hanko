import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { TodoClient, Todos } from "./TodoClient";
import styles from "./Todo.module.css";

const api = process.env.REACT_APP_TODO_API!;

function Todo() {
  const navigate = useNavigate();
  const [todos, setTodos] = useState<Todos>([]);
  const [description, setDescription] = useState<string>("");
  const [error, setError] = useState<Error | null>(null);
  const client = useMemo(() => new TodoClient(api), []);

  const addTodo = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const todo = { description, checked: false };

    client
      .addTodo(todo)
      .then((res) => {
        if (res.status === 401) {
          navigate("/");
          return;
        }

        setDescription("");
        listTodos();

        return;
      })
      .catch(setError);
  };

  const listTodos = useCallback(() => {
    client
      .listTodos()
      .then((res) => {
        if (res.status === 401) {
          navigate("/");
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
  }, [client, navigate]);

  const patchTodo = (id: string, checked: boolean) => {
    client
      .patchTodo(id, checked)
      .then((res) => {
        if (res.status === 401) {
          navigate("/");
          return;
        }

        listTodos();

        return;
      })
      .catch(setError);
  };

  const deleteTodo = (id: string) => {
    client
      .deleteTodo(id)
      .then((res) => {
        if (res.status === 401) {
          navigate("/");
          return;
        }

        listTodos();

        return;
      })
      .catch(setError);
  };

  const logout = () => {
    client
      .logout()
      .then(() => {
        navigate("/");
        return;
      })
      .catch(setError);
  };

  const profile = () => {
    navigate("/profile");
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
}

export default Todo;
