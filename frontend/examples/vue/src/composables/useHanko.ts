import { Hanko } from "@teamhanko/hanko-frontend-sdk";

export function useHanko() {
  const hankoAPI = import.meta.env.VITE_HANKO_API;
  return { hankoClient: new Hanko(hankoAPI) };
}
