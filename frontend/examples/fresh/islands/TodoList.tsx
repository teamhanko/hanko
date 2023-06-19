import { useEffect, useState, useRef } from "preact/hooks";
import { TodoItem } from "../components/TodoItem.tsx";

export default function TodoList() {
  const [description, setDescription] = useState("");
  const [todos, setTodos] = useState<Array<Todo>>([]);
  const [error, setError] = useState<string>("");
  const modalRef = useRef();

  useEffect(async () => {
    const response = await fetch("/api/todo");
    if (response.ok) {
      setTodos(await response.json());
    }
    if (response.status === 401) {
      setError(response.statusText);
    }
  }, []);

  useEffect(() => {
    if (error)
      modalRef.current.show()
  }, [error]);

  const handleSubmit = async (event: Event) => {
    event.preventDefault();

    if (description.trim().length) {
      const response = await fetch("/api/todo", {
        method: "POST",
        body: JSON.stringify({
          description,
          checked: false,
        })
      });

      if (response.ok) {
        const { todoID } = await response.json();
        setTodos([...todos, { todoID, description, checked: false }]);
        setDescription("");
      }
    }
  }

  const handleDelete = async (todoID: string) => {
    setTodos(todos.filter((todo) => todo.todoID !== todoID));
  };

  const handleUpdate = async (todoID: string, checked: boolean) => {
    setTodos(todos.map((todo) => {
      if (todo.todoID === todoID) {
        return { ...todo, checked };
      }
      return todo;
    }));
  };

  return (
    <>
      {
        error &&
        <dialog ref={modalRef} class="p-4">
          <div class="error">{error}</div>
          Please login again.<br /><br />
          <a href="/" class="px-8 py-1 bg-blue-500 text-white rounded">Login</a>
        </dialog>
      }
      <div class="mt-2">
        <div class="">
          <form class="flex" onSubmit={handleSubmit}>
            <input class="w-full mr-2 border-blue-500 rounded p-1.5 border-2 text-black" type="text" value={description} onChange={(event) => setDescription(event.target.value)} />
            <button class=" rounded border-blue-500 border-2! py-2 px-4">🆕</button>
          </form>
        </div>
        <div class="mt-8">
          {
            todos.map((todo) => <TodoItem key={todo.todoID} {...todo} onUpdate={handleUpdate} onDelete={handleDelete} onError={(error) => setError(error)} />)
          }
        </div>
      </div>
    </>
  );
}
