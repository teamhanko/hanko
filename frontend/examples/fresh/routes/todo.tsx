import LogoutButton from "../islands/LogoutButton.tsx";
import TodoList from "../islands/TodoList.tsx";

export default function Todo() {
  return (
    <>
      <nav class="nav flex justify-end gap-3">
        <button disabled class="button">Todos</button>
        <a href="/profile" class="button">Profile</a>
        <LogoutButton/>
      </nav>
      <TodoList/>
    </>
  );
}
