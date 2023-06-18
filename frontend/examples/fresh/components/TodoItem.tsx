import { useState } from "preact/hooks";

export function TodoItem(props: Todo) {
  const [checked, setChecked] = useState(props.checked);

  const updateTodo = async (event: Event) => {
    const target = event.target as HTMLInputElement;
    const response = await fetch(`/api/todo/${props.todoID}/`, {
      method: "PATCH",
      body: JSON.stringify({
        checked: target.checked,
      }),
    });
    setChecked(target.checked);
  };

  return (
    <div class="min-h-40 p-2 my-2 bg-white text-black rounded">
      <input class="mr-2" type="checkbox" value={checked} onChange={updateTodo}/>
      <span>{props.description}</span>
    </div>
  );
}
