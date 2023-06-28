import LogoutButton from "../islands/LogoutButton.tsx";
import TodoList from "../islands/TodoList.tsx";

export default function Todo() {
  return (
    <>
      <nav class="nav flex justify-end gap-3">
        <a href="#" class="button">Todos</a>
        <a href="/profile" class="button">Profile</a>
        <LogoutButton/>
      </nav>
      <TodoList/>
    </>
  );
}
