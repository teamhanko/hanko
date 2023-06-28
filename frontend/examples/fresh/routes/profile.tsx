import Profile from "../islands/Profile.tsx";
import LogoutButton from "../islands/LogoutButton.tsx";

export default function UserProfile() {
  return (
    <>
      <nav class="nav flex justify-end gap-3">
        <a href="/todo" class="button">Todos</a>
        <a href="/profile" class="button">Profile</a>
        <LogoutButton/>
      </nav>
      <Profile/>
    </>
  );
}
