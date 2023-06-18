import { useEffect, useState } from "preact/hooks";
import { TodoItem } from "../components/TodoItem.tsx";

export default function TodoList() {
  const [description, setDescription] = useState("");
  const [todos, setTodos] = useState<Array<Todo>>([]);

  useEffect(async () => {
    const response = await fetch("/api/todo");
    if (response.ok) {
      setTodos(await response.json());
    }
    if (response.status === 401) {
      alert("You're not authenticated.");
    }
  }, []);

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

  return (
    <div class="mt-2">
      <div class="">
        <form class="flex" onSubmit={handleSubmit}>
          <input class="w-full mr-2 border-blue-500 rounded p-1.5 border-2 text-black" type="text" value={description} onChange={(event) => setDescription(event.target.value)} />
          <button class=" rounded border-blue-500 border-2! py-2 px-4">🆕</button>
        </form>
      </div>
      <div class="mt-8">
        {
          todos.map((todo) => <TodoItem key={todo.todoID} {...todo} />)
        }
      </div>
    </div>
  );
}
