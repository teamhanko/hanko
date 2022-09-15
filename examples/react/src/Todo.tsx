import React, {useCallback, useEffect, useMemo, useState} from "react";
import {useNavigate} from "react-router-dom";
import {TodoClient, TodoList} from "./TodoClient";
import styles from "./Todo.module.css";

const api = process.env.REACT_APP_BACKEND!;

function Todo() {
  const navigate = useNavigate();
  const [todos, setTodos] = useState<TodoList>([]);
  const [description, setDescription] = useState<string>("");
  const [error, setError] = useState<Error | null>(null);
  const client = useMemo(() => new TodoClient(api), []);

  const addTodo = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    const entry = {description, checked: false};

    client.addTodo(entry).then((res) => {
      if (res.status === 401) {
        navigate("/");
        return;
      }

      setDescription("");
      listTodos();

      return;
    }).catch((e) => {
      setError(e);
    });
  }

  const listTodos = useCallback(() => {
    client.listTodos().then((res) => {
      if (res.status === 401) {
        navigate('/');
        return;
      }

      return res.json();
    }).then((t) => {
      if (t) {
        setTodos(t);
      }
    }).catch((e) => {
      setError(e);
    });
  }, [client, navigate]);

  const patchTodo = (id: number, checked: boolean) => {
    client.patchTodo(id, checked).then((res) => {
      if (res.status === 401) {
        navigate("/");
        return;
      }

      listTodos();

      return;
    }).catch((e) => {
      setError(e);
    });
  }

  const deleteTodo = (id: number) => {
    client.deleteTodo(id).then((res) => {
      if (res.status === 401) {
        navigate("/");
        return;
      }

      listTodos();

      return;
    }).catch((e) => {
      setError(e);
    });
  }

  const logout = () => {
    client.logout().then(() => {
      navigate('/');
      return;
    }).catch((e) => {
      setError(e);
    });
  }

  const changeDescription = (event: React.ChangeEvent<HTMLInputElement>) => {
    setDescription(event.currentTarget.value);
  }

  const changeCheckbox = (event: React.ChangeEvent<HTMLInputElement>) => {
    const {currentTarget} = event;
    patchTodo(Number(currentTarget.value), currentTarget.checked);
  }

  useEffect(() => {
    listTodos();
  }, [listTodos]);

  return <>
    <nav className={styles.nav}>
      <button onClick={logout} className={styles.button}>logout</button>
    </nav>
    <div className={styles.content}>
      <h1 className={styles.headline}>Todos</h1>
      <div className={styles.error}>{error?.message}</div>
      <form className={styles.form}>
        <input className={styles.input} type={"text"} value={description} onChange={changeDescription}/>
        <button onClick={addTodo} className={styles.button}>+</button>
      </form>
      <div className={styles.list}>
        {todos.map((t, id) => (
          <div className={styles.item} key={id}>
            <input className={styles.checkbox} id={id.toString(10)} type={"checkbox"} value={id}
                   checked={t.checked} onChange={changeCheckbox}/>
            <label className={styles.description} htmlFor={id.toString(10)}>{t.description}</label>
            <button className={styles.button} onClick={() => deleteTodo(id)}>×</button>
          </div>
        ))}
      </div>
    </div>
  </>;
}

export default Todo;
