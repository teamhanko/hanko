import { TodoClient } from "@/utils/TodoClient";

export function useTodo() {
  const todoAPI = import.meta.env.VITE_TODO_API;
  return { todoClient: new TodoClient(todoAPI) };
}
