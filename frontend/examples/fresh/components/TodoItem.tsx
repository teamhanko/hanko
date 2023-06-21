import { useState } from "preact/hooks";

export interface TodoItemProps extends Todo {
  onUpdate: (todoID: string, checked: boolean) => void;
  onDelete: (todoID: string) => void;
  onError: (error: string) => void;
}

export function TodoItem(props: TodoItemProps) {
  const [checked, setChecked] = useState(props.checked);

  const updateTodo = async (event: Event) => {
    const target = event.target as HTMLInputElement;
    const response = await fetch(`/api/todo/${props.todoID}/`, {
      method: "PATCH",
      body: JSON.stringify({
        checked: target.checked,
      }),
    });

    if (response.ok) {
      setChecked(target.checked);
      props.onUpdate(props.todoID, target.checked);
    }
    else {
      props.onError(response.statusText);
    }
  };

  const deleteTodo = async () => {
    const response = await fetch(`/api/todo/${props.todoID}/`, {
      method: "DELETE",
    });

    if (response.ok) {
      props.onDelete(props.todoID);
    }
    else {
      props.onError(response.statusText);
    }
  };

  return (
    <div class="min-h-40 p-2 my-2 bg-white text-black rounded flex gap-1 items-center">
      <input class="mr-2" type="checkbox" value={checked} checked={checked} onChange={updateTodo} />
      <span class="w-full">{props.description}</span>
      <button class="ml-2 rounded py-1 px-2" onClick={deleteTodo}>x</button>
    </div>
  );
}
